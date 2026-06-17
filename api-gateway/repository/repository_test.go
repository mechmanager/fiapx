package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/mechmanager/fiapx/api-gateway/domain"
	"github.com/mechmanager/fiapx/api-gateway/repository"
)

// --- Mock DB ---

type mockDB struct {
	execFn     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	queryFn    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	queryRowFn func(ctx context.Context, sql string, args ...any) pgx.Row
}

func (m *mockDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return m.execFn(ctx, sql, args...)
}
func (m *mockDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return m.queryFn(ctx, sql, args...)
}
func (m *mockDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return m.queryRowFn(ctx, sql, args...)
}

// --- Mock Row ---

type mockRow struct {
	scanFn func(dest ...any) error
}

func (r *mockRow) Scan(dest ...any) error { return r.scanFn(dest...) }

// --- Mock Rows ---

type mockRows struct {
	rows    [][]any
	current int
	err     error
}

func (r *mockRows) Next() bool {
	r.current++
	return r.current <= len(r.rows)
}
func (r *mockRows) Scan(dest ...any) error {
	row := r.rows[r.current-1]
	for i, d := range dest {
		switch p := d.(type) {
		case *uuid.UUID:
			*p = row[i].(uuid.UUID)
		case *string:
			*p = row[i].(string)
		case *int:
			*p = row[i].(int)
		case *time.Time:
			*p = row[i].(time.Time)
		}
	}
	return nil
}
func (r *mockRows) Err() error                                   { return r.err }
func (r *mockRows) Close()                                       {}
func (r *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *mockRows) Values() ([]any, error)                       { return nil, nil }
func (r *mockRows) RawValues() [][]byte                          { return nil }
func (r *mockRows) Conn() *pgx.Conn                              { return nil }

// --- UserRepo tests ---

func TestUserRepo_Create_OK(t *testing.T) {
	db := &mockDB{
		execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("INSERT 1"), nil
		},
	}
	repo := repository.NewUserRepo(db)
	err := repo.Create(context.Background(), &domain.User{
		ID: uuid.New(), Name: "Alice", Email: "a@b.com", PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
}

func TestUserRepo_Create_Error(t *testing.T) {
	db := &mockDB{
		execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, errors.New("constraint violation")
		},
	}
	repo := repository.NewUserRepo(db)
	err := repo.Create(context.Background(), &domain.User{ID: uuid.New()})
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestUserRepo_FindByEmail_OK(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	db := &mockDB{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{
				scanFn: func(dest ...any) error {
					*(dest[0].(*uuid.UUID)) = id
					*(dest[1].(*string)) = "Alice"
					*(dest[2].(*string)) = "a@b.com"
					*(dest[3].(*string)) = "hash"
					*(dest[4].(*time.Time)) = now
					return nil
				},
			}
		},
	}
	repo := repository.NewUserRepo(db)
	user, err := repo.FindByEmail(context.Background(), "a@b.com")
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if user.ID != id {
		t.Errorf("ID incorreto: %v", user.ID)
	}
}

func TestUserRepo_FindByEmail_NotFound(t *testing.T) {
	db := &mockDB{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{
				scanFn: func(_ ...any) error { return pgx.ErrNoRows },
			}
		},
	}
	repo := repository.NewUserRepo(db)
	_, err := repo.FindByEmail(context.Background(), "nao@existe.com")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("esperava ErrUserNotFound, obteve %v", err)
	}
}

func TestUserRepo_FindByEmail_DBError(t *testing.T) {
	db := &mockDB{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{
				scanFn: func(_ ...any) error { return errors.New("timeout") },
			}
		},
	}
	repo := repository.NewUserRepo(db)
	_, err := repo.FindByEmail(context.Background(), "a@b.com")
	if err == nil {
		t.Fatal("esperava erro de banco")
	}
}

// --- VideoRepo tests ---

func TestVideoRepo_Create_OK(t *testing.T) {
	db := &mockDB{
		execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
			return pgconn.NewCommandTag("INSERT 1"), nil
		},
	}
	repo := repository.NewVideoRepo(db)
	err := repo.Create(context.Background(), &domain.Video{
		ID: uuid.New(), UserID: uuid.New(),
		OriginalFilename: "v.mp4", S3Key: "k", Status: domain.StatusPending,
	})
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
}

func TestVideoRepo_ListByUser_OK(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()

	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return &mockRows{
				rows: [][]any{
					{id, userID, "v.mp4", "keys3", "PENDING", "", "", 0, now, now},
				},
			}, nil
		},
	}
	repo := repository.NewVideoRepo(db)
	videos, err := repo.ListByUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if len(videos) != 1 {
		t.Fatalf("esperava 1 vídeo, obteve %d", len(videos))
	}
	if videos[0].ID != id {
		t.Errorf("ID incorreto: %v", videos[0].ID)
	}
}

func TestVideoRepo_ListByUser_QueryError(t *testing.T) {
	db := &mockDB{
		queryFn: func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
			return nil, errors.New("timeout")
		},
	}
	repo := repository.NewVideoRepo(db)
	_, err := repo.ListByUser(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestVideoRepo_FindByID_OK(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()
	db := &mockDB{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{
				scanFn: func(dest ...any) error {
					*(dest[0].(*uuid.UUID)) = id
					*(dest[1].(*uuid.UUID)) = userID
					*(dest[2].(*string)) = "v.mp4"
					*(dest[3].(*string)) = "key"
					*(dest[4].(*string)) = "DONE"
					*(dest[5].(*string)) = ""
					*(dest[6].(*string)) = "zip/key"
					*(dest[7].(*int)) = 10
					*(dest[8].(*time.Time)) = now
					*(dest[9].(*time.Time)) = now
					return nil
				},
			}
		},
	}
	repo := repository.NewVideoRepo(db)
	v, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if v.Status != "DONE" {
		t.Errorf("status incorreto: %q", v.Status)
	}
}

func TestVideoRepo_FindByID_NotFound(t *testing.T) {
	db := &mockDB{
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return &mockRow{scanFn: func(_ ...any) error { return pgx.ErrNoRows }}
		},
	}
	repo := repository.NewVideoRepo(db)
	_, err := repo.FindByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrVideoNotFound) {
		t.Fatalf("esperava ErrVideoNotFound, obteve %v", err)
	}
}
