package practice

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nguyensongtai/lingora-api/internal/db"
	"github.com/nguyensongtai/lingora-api/internal/platform/postgres"
)

// Repo lưu điểm luyện tập trong bài — thứ duy nhất gói này tự lưu. Câu hỏi và
// lịch ôn vẫn không có bảng nào ở đây.
type Repo struct {
	q *db.Queries
}

// NewRepo nhận DBTX nên dùng được cả với pool lẫn transaction.
func NewRepo(dbtx db.DBTX) *Repo {
	return &Repo{q: db.New(dbtx)}
}

// RecordScore ghi một lượt và trả về điểm tốt nhất sau khi ghi.
func (r *Repo) RecordScore(ctx context.Context, userID, lessonID string, correct, total int) (Score, error) {
	user, err := parseID(userID)
	if err != nil {
		return Score{}, err
	}
	lesson, err := parseID(lessonID)
	if err != nil {
		return Score{}, err
	}

	row, err := r.q.RecordLessonPracticeScore(ctx, db.RecordLessonPracticeScoreParams{
		UserID:   user,
		LessonID: lesson,
		Correct:  int32(correct),
		Total:    int32(total),
	})
	if err != nil {
		return Score{}, fmt.Errorf("record practice score of %s: %w", lessonID, err)
	}
	return Score{
		LessonID:    postgres.UUIDString(row.LessonID),
		BestCorrect: int(row.BestCorrect),
		Total:       int(row.Total),
		Attempts:    int(row.Attempts),
	}, nil
}

// ListScores trả về điểm của một người ở mọi bài còn sống.
func (r *Repo) ListScores(ctx context.Context, userID string) ([]Score, error) {
	user, err := parseID(userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.q.ListLessonPracticeScores(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("list practice scores of %s: %w", userID, err)
	}
	scores := make([]Score, 0, len(rows))
	for _, row := range rows {
		scores = append(scores, Score{
			LessonID:    postgres.UUIDString(row.LessonID),
			BestCorrect: int(row.BestCorrect),
			Total:       int(row.Total),
			Attempts:    int(row.Attempts),
		})
	}
	return scores, nil
}

func parseID(raw string) (pgtype.UUID, error) {
	id, err := postgres.ParseUUID(raw)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("%w: %q", ErrInvalidID, raw)
	}
	return id, nil
}
