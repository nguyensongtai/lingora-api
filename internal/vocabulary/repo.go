package vocabulary

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nguyensongtai/lingora-api/internal/db"
	"github.com/nguyensongtai/lingora-api/internal/platform/postgres"
)

const uniqueViolation = "23505"

// Repo đọc ghi từ vựng và lịch ôn trên PostgreSQL.
type Repo struct {
	q *db.Queries
}

// NewRepo nhận DBTX nên dùng được cả với pool lẫn transaction.
func NewRepo(dbtx db.DBTX) *Repo {
	return &Repo{q: db.New(dbtx)}
}

// Create thêm một từ vào cuối bài.
func (r *Repo) Create(ctx context.Context, lessonID string, params CreateParams) (Entry, error) {
	id, err := parseID(lessonID)
	if err != nil {
		return Entry{}, err
	}

	row, err := r.q.CreateVocabularyEntry(ctx, db.CreateVocabularyEntryParams{
		LessonID:  id,
		Word:      params.Word,
		Ipa:       params.IPA,
		Meaning:   params.Meaning,
		Example:   params.Example,
		ExampleVi: params.ExampleVI,
	})
	if err != nil {
		if isUniqueViolation(err, "vocabulary_entries_lesson_word_key") {
			return Entry{}, fmt.Errorf("create vocabulary %q: %w", params.Word, ErrWordTaken)
		}
		return Entry{}, fmt.Errorf("create vocabulary: %w", err)
	}
	return toEntry(row), nil
}

// Get đọc một từ chưa bị xoá mềm.
func (r *Repo) Get(ctx context.Context, entryID string) (Entry, error) {
	id, err := parseID(entryID)
	if err != nil {
		return Entry{}, err
	}

	row, err := r.q.GetVocabularyEntry(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Entry{}, fmt.Errorf("get vocabulary %s: %w", entryID, ErrNotFound)
		}
		return Entry{}, fmt.Errorf("get vocabulary %s: %w", entryID, err)
	}
	return toEntry(row), nil
}

// ListByLesson trả về từ của một bài theo đúng thứ tự position.
func (r *Repo) ListByLesson(ctx context.Context, lessonID string) ([]Entry, error) {
	id, err := parseID(lessonID)
	if err != nil {
		return nil, err
	}

	rows, err := r.q.ListVocabularyByLesson(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list vocabulary of lesson %s: %w", lessonID, err)
	}

	entries := make([]Entry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, toEntry(row))
	}
	return entries, nil
}

// Update áp dụng partial update lên một từ.
func (r *Repo) Update(ctx context.Context, entryID string, params UpdateParams) (Entry, error) {
	id, err := parseID(entryID)
	if err != nil {
		return Entry{}, err
	}

	row, err := r.q.UpdateVocabularyEntry(ctx, db.UpdateVocabularyEntryParams{
		ID:        id,
		Word:      params.Word,
		Ipa:       params.IPA,
		Meaning:   params.Meaning,
		Example:   params.Example,
		ExampleVi: params.ExampleVI,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Entry{}, fmt.Errorf("update vocabulary %s: %w", entryID, ErrNotFound)
		}
		if isUniqueViolation(err, "vocabulary_entries_lesson_word_key") {
			return Entry{}, fmt.Errorf("update vocabulary %s: %w", entryID, ErrWordTaken)
		}
		return Entry{}, fmt.Errorf("update vocabulary %s: %w", entryID, err)
	}
	return toEntry(row), nil
}

// SoftDelete đánh dấu deleted_at cho một từ.
func (r *Repo) SoftDelete(ctx context.Context, entryID string) error {
	id, err := parseID(entryID)
	if err != nil {
		return err
	}

	affected, err := r.q.SoftDeleteVocabularyEntry(ctx, id)
	if err != nil {
		return fmt.Errorf("soft delete vocabulary %s: %w", entryID, err)
	}
	if affected == 0 {
		return fmt.Errorf("soft delete vocabulary %s: %w", entryID, ErrNotFound)
	}
	return nil
}

// ListForUser trả về mọi từ đã mở khoá của một người, kèm lịch ôn nếu đã có.
func (r *Repo) ListForUser(ctx context.Context, userID string) ([]Card, error) {
	id, err := parseID(userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.q.ListVocabularyForUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list vocabulary for %s: %w", userID, err)
	}

	cards := make([]Card, 0, len(rows))
	for _, row := range rows {
		card := Card{
			Entry: Entry{
				ID:        postgres.UUIDString(row.ID),
				LessonID:  postgres.UUIDString(row.LessonID),
				Word:      row.Word,
				IPA:       row.Ipa,
				Meaning:   row.Meaning,
				Example:   row.Example,
				ExampleVI: row.ExampleVi,
				Position:  row.Position,
				CreatedAt: row.CreatedAt,
				UpdatedAt: row.UpdatedAt,
			},
			Level: string(row.Level),
		}
		// Bốn cột của LEFT JOIN cùng nil hoặc cùng có giá trị; kiểm due_on là đủ
		// vì nó NOT NULL trong bảng.
		if row.DueOn.Valid {
			card.Review = &Review{
				EaseFactor:     deref(row.EaseFactor),
				IntervalDays:   deref(row.IntervalDays),
				Repetitions:    deref(row.Repetitions),
				DueOn:          row.DueOn.Time,
				LastReviewedAt: deref(row.LastReviewedAt),
			}
		}
		cards = append(cards, card)
	}
	return cards, nil
}

// SaveReview ghi trạng thái SM-2 mới, tạo hàng nếu đây là lần ôn đầu.
func (r *Repo) SaveReview(ctx context.Context, userID, entryID string, review Review) (Review, error) {
	user, err := parseID(userID)
	if err != nil {
		return Review{}, err
	}
	entry, err := parseID(entryID)
	if err != nil {
		return Review{}, err
	}

	row, err := r.q.UpsertVocabularyReview(ctx, db.UpsertVocabularyReviewParams{
		UserID:       user,
		EntryID:      entry,
		EaseFactor:   review.EaseFactor,
		IntervalDays: review.IntervalDays,
		Repetitions:  review.Repetitions,
		DueOn:        pgtype.Date{Time: review.DueOn, Valid: true},
	})
	if err != nil {
		return Review{}, fmt.Errorf("save review of %s: %w", entryID, err)
	}

	return Review{
		EaseFactor:     row.EaseFactor,
		IntervalDays:   row.IntervalDays,
		Repetitions:    row.Repetitions,
		DueOn:          row.DueOn.Time,
		LastReviewedAt: row.LastReviewedAt,
	}, nil
}

// CountNewThisWeek đếm số từ được ôn lần đầu trong bảy ngày qua.
func (r *Repo) CountNewThisWeek(ctx context.Context, userID string) (int64, error) {
	id, err := parseID(userID)
	if err != nil {
		return 0, err
	}

	count, err := r.q.CountVocabularyLearnedThisWeek(ctx, id)
	if err != nil {
		return 0, fmt.Errorf("count new vocabulary of %s: %w", userID, err)
	}
	return count, nil
}

func parseID(raw string) (pgtype.UUID, error) {
	id, err := postgres.ParseUUID(raw)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("%w: %q", ErrInvalidID, raw)
	}
	return id, nil
}

func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == uniqueViolation && pgErr.ConstraintName == constraint
}

// deref trả về giá trị zero khi con trỏ nil; dùng cho các cột của LEFT JOIN.
func deref[T any](value *T) T {
	if value == nil {
		var zero T
		return zero
	}
	return *value
}

func toEntry(row db.VocabularyEntry) Entry {
	return Entry{
		ID:        postgres.UUIDString(row.ID),
		LessonID:  postgres.UUIDString(row.LessonID),
		Word:      row.Word,
		IPA:       row.Ipa,
		Meaning:   row.Meaning,
		Example:   row.Example,
		ExampleVI: row.ExampleVi,
		Position:  row.Position,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
