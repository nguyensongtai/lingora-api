package postgres_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nguyensongtai/lingora-api/internal/platform/postgres"
)

func TestParseUUIDRoundTrip(t *testing.T) {
	t.Parallel()

	const want = "018f3a9c-7b2e-7c31-9a55-0f1d2e3a4b5c"

	id, err := postgres.ParseUUID(want)
	if err != nil {
		t.Fatalf("ParseUUID(%q) returned error: %v", want, err)
	}
	if !id.Valid {
		t.Fatalf("ParseUUID(%q) produced an invalid uuid", want)
	}
	if got := postgres.UUIDString(id); got != want {
		t.Fatalf("UUIDString() = %q, want %q", got, want)
	}
}

func TestParseUUIDRejectsMalformedInput(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"", "not-a-uuid", "018f3a9c7b2e7c319a550f1d2e3a4b"} {
		if _, err := postgres.ParseUUID(raw); err == nil {
			t.Errorf("ParseUUID(%q) = nil error, want error", raw)
		}
	}
}

func TestUUIDStringOfNullUUIDIsEmpty(t *testing.T) {
	t.Parallel()

	if got := postgres.UUIDString(pgtype.UUID{}); got != "" {
		t.Fatalf("UUIDString(null) = %q, want empty string", got)
	}
}
