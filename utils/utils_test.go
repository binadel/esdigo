package utils

import (
	"regexp"
	"sync"
	"testing"
)

// TestUnsafeRoundTrip: UnsafeBytes and UnsafeString are inverses and agree with the
// safe conversions for arbitrary byte content (they only reinterpret the header).
func TestUnsafeRoundTrip(t *testing.T) {
	cases := []string{"a", "hello world", "unicode: café ☃😀", string([]byte{0, 1, 2, 255})}
	for _, s := range cases {
		b := UnsafeBytes(s)
		if string(b) != s {
			t.Errorf("UnsafeBytes(%q) -> %q", s, string(b))
		}
		if got := UnsafeString(b); got != s {
			t.Errorf("UnsafeString(UnsafeBytes(%q)) = %q", s, got)
		}
		if UnsafeString([]byte(s)) != s {
			t.Errorf("UnsafeString differs from string() for %q", s)
		}
	}
}

// TestUnsafeEmpty: the empty inputs take the nil/"" fast paths.
func TestUnsafeEmpty(t *testing.T) {
	if got := UnsafeString(nil); got != "" {
		t.Errorf("UnsafeString(nil) = %q, want empty", got)
	}
	if got := UnsafeString([]byte{}); got != "" {
		t.Errorf("UnsafeString([]byte{}) = %q, want empty", got)
	}
	if b := UnsafeBytes(""); b != nil {
		t.Errorf("UnsafeBytes(\"\") = %v, want nil", b)
	}
}

// TestRegexCacheGet: Get compiles a working regex and returns the SAME cached
// *regexp.Regexp on a hit; a different pattern compiles a distinct one.
func TestRegexCacheGet(t *testing.T) {
	var c RegexCache
	re1, err := c.Get(`^[a-z]+$`)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if re1 == nil || !re1.MatchString("abc") || re1.MatchString("ABC") {
		t.Errorf("compiled regex behaves incorrectly")
	}
	re2, _ := c.Get(`^[a-z]+$`)
	if re1 != re2 {
		t.Errorf("Get should return the cached *regexp.Regexp (same pointer)")
	}
	if re3, _ := c.Get(`\d+`); re3 == re1 {
		t.Errorf("a different pattern should compile a different regex")
	}
}

// TestRegexCacheInvalid: an invalid pattern returns an error from Get and panics
// from MustGet.
func TestRegexCacheInvalid(t *testing.T) {
	var c RegexCache
	if _, err := c.Get(`(`); err == nil {
		t.Errorf("an invalid pattern should return an error")
	}
	defer func() {
		if recover() == nil {
			t.Errorf("MustGet should panic on an invalid pattern")
		}
	}()
	c.MustGet(`(`)
}

// TestRegexCacheConcurrent: many goroutines racing on the same uncached pattern all
// receive the identical compiled regex (the leader/follower path). Run with -race.
func TestRegexCacheConcurrent(t *testing.T) {
	var c RegexCache
	const pattern = `^x[0-9]{3}$`

	var wg sync.WaitGroup
	results := make([]*regexp.Regexp, 32)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			re, err := c.Get(pattern)
			if err != nil {
				t.Errorf("Get: %v", err)
			}
			results[i] = re
		}(i)
	}
	wg.Wait()

	for i, re := range results {
		if re == nil || re != results[0] {
			t.Errorf("result %d = %p; all concurrent Gets must return the same cached regex %p", i, re, results[0])
		}
	}
	if !results[0].MatchString("x123") || results[0].MatchString("x12") {
		t.Errorf("cached regex matches incorrectly")
	}
}
