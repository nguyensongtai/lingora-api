// Command api is the HTTP entrypoint of the Lingora backend.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/nguyensongtai/lingora-api/internal/auth"
	"github.com/nguyensongtai/lingora-api/internal/config"
	"github.com/nguyensongtai/lingora-api/internal/course"
	"github.com/nguyensongtai/lingora-api/internal/httpx"
	"github.com/nguyensongtai/lingora-api/internal/platform/postgres"
	"github.com/nguyensongtai/lingora-api/internal/progress"
	"github.com/nguyensongtai/lingora-api/internal/ratelimit"
	"github.com/nguyensongtai/lingora-api/internal/user"
	"github.com/nguyensongtai/lingora-api/internal/vocabulary"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api exited with error", slog.Any("error", err))
		os.Exit(1)
	}
}

// run wires every dependency by hand and blocks until the process is signalled.
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	})))

	verifier, err := auth.NewVerifier(cfg.JWTSecret)
	if err != nil {
		return fmt.Errorf("build token verifier: %w", err)
	}

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	// Dependency được nối tay theo mạch repo -> service -> handler.
	// Một limiter dùng chung cho cả hạn mức theo IP lẫn theo email.
	limiter := ratelimit.NewMemory()

	authConfig := auth.Config{
		Limiter:    limiter,
		Secret:     cfg.JWTSecret,
		AccessTTL:  cfg.AccessTokenTTL,
		RefreshTTL: cfg.RefreshTokenTTL,
	}
	// Thiếu client id hoặc secret thì để nil: endpoint Google trả 503 chứ không
	// chạy nửa vời với cấu hình rỗng.
	googleConfig := auth.GoogleConfig{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
	}
	if googleConfig.Configured() {
		authConfig.Google = auth.NewGoogleClient(googleConfig)
	} else {
		slog.Warn("google sign-in disabled: GOOGLE_CLIENT_ID/GOOGLE_CLIENT_SECRET chưa được đặt")
	}

	authService, err := auth.NewService(user.NewRepo(pool), auth.NewSessionRepo(pool), authConfig)
	if err != nil {
		return fmt.Errorf("build auth service: %w", err)
	}

	authHandler := auth.NewHandler(authService, limiter)
	courseRepo := course.NewRepo(pool)
	courseHandler := course.NewHandler(course.NewService(courseRepo))
	// progress hỏi course xem bài còn sống hay không, nên dùng chung đúng repo đó.
	progressHandler := progress.NewHandler(progress.NewService(progress.NewRepo(pool), courseRepo))
	vocabularyHandler := vocabulary.NewHandler(vocabulary.NewService(vocabulary.NewRepo(pool), courseRepo))

	router := chi.NewRouter()
	router.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Recoverer,
		middleware.Timeout(30*time.Second),
		httpx.CORS(cfg.CORSAllowedOrigins),
	)

	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	router.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			slog.ErrorContext(r.Context(), "readiness check failed", slog.Any("error", err))
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "degraded"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	router.Route("/v1", func(r chi.Router) {
		authHandler.Mount(r, verifier.RequireAuthenticated())
		courseHandler.Mount(r, verifier.RequireRole(auth.RoleAdmin))
		progressHandler.Mount(r, verifier.RequireAuthenticated())
		vocabularyHandler.Mount(r, verifier.RequireRole(auth.RoleAdmin), verifier.RequireAuthenticated())
	})

	srv := &http.Server{
		Addr:              net.JoinHostPort("", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("api listening", slog.String("addr", srv.Addr), slog.String("env", cfg.Env))
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %w", err)
		}
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	slog.Info("api stopped")
	return nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("encode response body", slog.Any("error", err))
	}
}
