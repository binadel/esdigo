package utils

import (
	"regexp"
	"sync"
)

// entry wraps the compiled regex and a wait group to guarantee single compilation.
type entry struct {
	wg  sync.WaitGroup
	re  *regexp.Regexp
	err error
}

// RegexCache is a strictly zero-allocation (on read) append-only cache.
type RegexCache struct {
	items sync.Map
}

// Get returns a compiled regex. It is lock-free on cache hits.
func (c *RegexCache) Get(pattern string) (*regexp.Regexp, error) {
	if actual, ok := c.items.Load(pattern); ok {
		ent := actual.(*entry)

		// If another goroutine just added this but is still compiling,
		// Wait() will safely block. If it's done, Wait() returns instantly.
		ent.wg.Wait()
		return ent.re, ent.err
	}

	ent := &entry{}
	ent.wg.Add(1) // Block followers before putting it in the map

	actual, loaded := c.items.LoadOrStore(pattern, ent)
	entActual := actual.(*entry)

	if !loaded {
		// Leader Goroutine: Compiles the regex
		entActual.re, entActual.err = regexp.Compile(pattern)
		entActual.wg.Done() // Unblock all waiting followers
	} else {
		// Follower Goroutine: Caught in a race, wait for the leader
		entActual.wg.Wait()
	}

	return entActual.re, entActual.err
}

// MustGet wraps Get and panics on invalid regex syntax.
func (c *RegexCache) MustGet(pattern string) *regexp.Regexp {
	re, err := c.Get(pattern)
	if err != nil {
		panic("regex cache: invalid developer-defined pattern - " + err.Error())
	}
	return re
}
