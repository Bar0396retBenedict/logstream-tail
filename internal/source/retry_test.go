package source

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryer_SucceedsImmediately(t *testing.T) {
	r := NewRetryer(DefaultRetryConfig())
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRetryer_RetriesUntilSuccess(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 0, BaseDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond}
	r := NewRetryer(cfg)
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		if calls < 3 {
			return errors.New("not yet")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRetryer_MaxAttempts(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond}
	r := NewRetryer(cfg)
	calls := 0
	sentinel := errors.New("always fail")
	err := r.Do(context.Background(), func() error {
		calls++
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRetryer_ContextCancellation(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 0, BaseDelay: 50 * time.Millisecond, MaxDelay: time.Second}
	r := NewRetryer(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	calls := 0
	err := r.Do(ctx, func() error {
		calls++
		return errors.New("fail")
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	if calls == 0 {
		t.Fatal("expected at least one call")
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	if cfg.MaxAttempts != 0 {
		t.Errorf("expected MaxAttempts 0, got %d", cfg.MaxAttempts)
	}
	if cfg.BaseDelay != 500*time.Millisecond {
		t.Errorf("unexpected BaseDelay: %v", cfg.BaseDelay)
	}
	if cfg.MaxDelay != 30*time.Second {
		t.Errorf("unexpected MaxDelay: %v", cfg.MaxDelay)
	}
}
