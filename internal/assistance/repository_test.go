package assistance

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// testPool connects to the database in TEST_DATABASE_URL, skipping when unset
// so unit runs stay hermetic.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Skipf("no test database: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func testUser(t *testing.T, pool *pgxpool.Pool, role string) string {
	t.Helper()
	id := uuid.NewString()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO users (id, phone_hash, phone_last4, role) VALUES ($1, $2, '0000', $3)`,
		id, "testhash-"+uuid.NewString(), role)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	})
	return id
}

func TestAcceptRequest_AtomicRace(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewRepository(pool)
	citizen := testUser(t, pool, "citizen")
	q, err := repo.CreateRequest(ctx, citizen, nil, nil, "hi", "application_help", "race", true, false)
	if err != nil {
		t.Fatal(err)
	}
	const racers = 10
	var wins int64
	var wg sync.WaitGroup
	for i := 0; i < racers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			helper := testUser(t, pool, "helper")
			if _, err := repo.AcceptRequest(ctx, q.ID, helper); err == nil {
				atomic.AddInt64(&wins, 1)
			} else if err != ErrConflict {
				t.Errorf("racer %d: unexpected error %v", i, err)
			}
		}(i)
	}
	wg.Wait()
	if wins != 1 {
		t.Fatalf("exactly one accepter must win, got %d", wins)
	}
	got, err := repo.GetRequest(ctx, q.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusAccepted || got.MatchedHelperID == nil {
		t.Fatalf("request not assigned: %+v", got)
	}
}

func TestRequestLifecycle_Transitions(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewRepository(pool)
	citizen := testUser(t, pool, "citizen")
	helper := testUser(t, pool, "helper")

	q, err := repo.CreateRequest(ctx, citizen, nil, nil, "ta", "document_help", "", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if q.Status != StatusPending {
		t.Fatalf("new request must be pending, got %s", q.Status)
	}
	if _, err := repo.AcceptRequest(ctx, q.ID, helper); err != nil {
		t.Fatal(err)
	}
	// Second accept must conflict.
	if _, err := repo.AcceptRequest(ctx, q.ID, testUser(t, pool, "helper")); err != ErrConflict {
		t.Fatalf("double accept must conflict, got %v", err)
	}
	// Illegal jump pending->completed is rejected by the allowed-from list.
	if _, err := repo.SetRequestStatus(ctx, q.ID, StatusCompleted, []string{StatusPending}); err != ErrConflict {
		t.Fatalf("illegal transition must conflict, got %v", err)
	}
	out, err := repo.SetRequestStatus(ctx, q.ID, StatusInSession, []string{StatusAccepted})
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != StatusInSession {
		t.Fatalf("want in_session, got %s", out.Status)
	}
	if _, err := repo.SetRequestStatus(ctx, q.ID, StatusCompleted, []string{StatusAccepted, StatusInSession}); err != nil {
		t.Fatal(err)
	}
	// Terminal states are final.
	if _, err := repo.SetRequestStatus(ctx, q.ID, StatusCancelled, []string{StatusPending}); err != ErrConflict {
		t.Fatalf("terminal state must be final, got %v", err)
	}
}

func TestMessageIdempotency(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewRepository(pool)
	citizen := testUser(t, pool, "citizen")
	helper := testUser(t, pool, "helper")
	q, err := repo.CreateRequest(ctx, citizen, nil, nil, "hi", "application_help", "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.AcceptRequest(ctx, q.ID, helper); err != nil {
		t.Fatal(err)
	}
	sess, err := repo.CreateSession(ctx, q.ID, helper, citizen, "application_help")
	if err != nil {
		t.Fatal(err)
	}
	m1, err := repo.AddMessage(ctx, sess.ID, citizen, "citizen", "dup-1", "hello")
	if err != nil {
		t.Fatal(err)
	}
	m2, err := repo.AddMessage(ctx, sess.ID, citizen, "citizen", "dup-1", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if m1.ID != m2.ID {
		t.Fatal("duplicate client_msg_id must return the original row")
	}
	list, err := repo.ListMessages(ctx, sess.ID, m1.CreatedAt.Add(-1), 50)
	if err != nil {
		t.Fatal(err)
	}
	_ = list
}
