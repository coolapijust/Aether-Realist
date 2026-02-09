package core

import (
	"sync"
	"time"
)

const DefaultReplayWindow = 30 * time.Second

// ReplayCache tracks recently seen IVs to prevent replay.
type ReplayCache struct {
	ttl   time.Duration
	cache sync.Map
	stop  chan struct{}
}

// NewReplayCache creates a replay cache with a cleanup loop.
func NewReplayCache(ttl time.Duration) *ReplayCache {
	rc := &ReplayCache{
		ttl:  ttl,
		stop: make(chan struct{}),
	}
	go rc.cleanupLoop()
	return rc
}

// SeenOrAdd returns true if the IV was already seen; otherwise stores it.
func (rc *ReplayCache) SeenOrAdd(iv []byte, now time.Time) bool {
	if len(iv) != headerIVLength {
		return true
	}
	key := string(append([]byte(nil), iv...))
	if _, loaded := rc.cache.LoadOrStore(key, now.Add(rc.ttl)); loaded {
		return true
	}
	return false
}

// Close stops the cleanup loop.
func (rc *ReplayCache) Close() {
	close(rc.stop)
}

func (rc *ReplayCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-rc.stop:
			return
		case now := <-ticker.C:
			rc.cache.Range(func(key, value any) bool {
				expiry, ok := value.(time.Time)
				if !ok || now.After(expiry) {
					rc.cache.Delete(key)
				}
				return true
			})
		}
	}
}
