package datasource

import (
	"context"
	"testing"
	"time"
)

func TestCacheSetGet(t *testing.T) {
	c := NewCache(1 * time.Second)

	// Set a value.
	c.Set("key1", "value1")
	v, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if v != "value1" {
		t.Fatalf("got %v, want value1", v)
	}
}

func TestCacheMiss(t *testing.T) {
	c := NewCache(1 * time.Second)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss for nonexistent key")
	}
}

func TestCacheExpiry(t *testing.T) {
	c := NewCache(1 * time.Millisecond)
	c.Set("key", "val")

	// Wait for expiry.
	time.Sleep(5 * time.Millisecond)
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected cache miss after TTL expiry")
	}
}

func TestCacheSetWithTTL(t *testing.T) {
	c := NewCache(1 * time.Hour) // default long TTL.
	c.SetWithTTL("quick", "val", 1*time.Millisecond)

	time.Sleep(5 * time.Millisecond)
	_, ok := c.Get("quick")
	if ok {
		t.Fatal("expected cache miss after custom TTL expiry")
	}
}

func TestCacheInvalidate(t *testing.T) {
	c := NewCache(1 * time.Hour)
	c.Set("key", "val")
	c.Invalidate("key")
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected cache miss after invalidation")
	}
}

func TestCacheFlush(t *testing.T) {
	c := NewCache(1 * time.Hour)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Flush()
	_, okA := c.Get("a")
	_, okB := c.Get("b")
	if okA || okB {
		t.Fatal("expected all entries flushed")
	}
}

func TestCacheCleanup(t *testing.T) {
	c := NewCache(1 * time.Millisecond)
	c.Set("expired", "val")
	time.Sleep(5 * time.Millisecond)

	c.Set("fresh", "val2")
	c.Cleanup()

	_, okExpired := c.Get("expired")
	_, okFresh := c.Get("fresh")
	if okExpired {
		t.Fatal("expected expired entry to be cleaned up")
	}
	if !okFresh {
		t.Fatal("expected fresh entry to survive cleanup")
	}
}

func TestRateLimiterAllowsBurst(t *testing.T) {
	rl := NewRateLimiter(3, time.Second)
	ctx := context.Background()

	// Should allow 3 immediate calls.
	for i := 0; i < 3; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Fatalf("Wait() #%d failed: %v", i, err)
		}
	}
}

func TestRateLimiterCancelledContext(t *testing.T) {
	rl := NewRateLimiter(1, time.Hour) // 1 token, very slow refill.
	ctx := context.Background()

	// Use the single token.
	if err := rl.Wait(ctx); err != nil {
		t.Fatalf("first Wait() failed: %v", err)
	}

	// Next call with cancelled context should fail.
	ctx2, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := rl.Wait(ctx2)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestErrHTTPError(t *testing.T) {
	e := &ErrHTTP{StatusCode: 404, Status: "404 Not Found", Body: "page not found"}
	msg := e.Error()
	if msg != "HTTP 404 404 Not Found: page not found" {
		t.Fatalf("unexpected error message: %s", msg)
	}
}

func TestCoalesce(t *testing.T) {
	tests := []struct {
		input []string
		want  string
	}{
		{[]string{"", "", "hello"}, "hello"},
		{[]string{"first", "second"}, "first"},
		{[]string{"", ""}, ""},
		{[]string{"  ", "actual"}, "actual"},
	}
	for _, tt := range tests {
		got := coalesce(tt.input...)
		if got != tt.want {
			t.Errorf("coalesce(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s    string
		subs []string
		want bool
	}{
		{"FII/FPI", []string{"FII", "FPI"}, true},
		{"DII", []string{"DII"}, true},
		{"FII/FPI", []string{"DII"}, false},
		{"hello world", []string{"HELLO"}, true}, // case-insensitive
		{"", []string{"a"}, false},
	}
	for _, tt := range tests {
		got := contains(tt.s, tt.subs...)
		if got != tt.want {
			t.Errorf("contains(%q, %v) = %v, want %v", tt.s, tt.subs, got, tt.want)
		}
	}
}

func TestParseScreenerNumber(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"1,234.56", 1234.56},
		{"₹1,234.56", 1234.56},
		{"12.5%", 12.5},
		{"10Cr", 1e8},
		{"5 Cr.", 5e7},
		{"2.5 Lakh", 2.5e5},
		{"", 0},
		{"N/A", 0},
	}
	for _, tt := range tests {
		got := parseScreenerNumber(tt.input)
		if got != tt.want {
			t.Errorf("parseScreenerNumber(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
