package server

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   int       //max token
	tokens     int       //mevcut token
	refillRate int       //saniyede kaç token ekleyebiliriz
	lastRefill time.Time //en sonki refill zamanı
	mu         sync.Mutex
}

// Yeni bucket oluştur
func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Token al
func (tb *TokenBucket) Take() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	tokensToAdd := int(elapsed.Seconds() * float64(tb.refillRate))

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}
}

// mevcut token sayısı
func (tb *TokenBucket) Available() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	return tb.tokens
}

type RateLimiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex

	defaultCapacity   int
	defaultRefillRate int
}

// yeni rate limiter oluştur
func NewRateLimiter(capacity, refillRate int) *RateLimiter {
	rl := &RateLimiter{
		buckets:           make(map[string]*TokenBucket),
		defaultCapacity:   capacity,
		defaultRefillRate: refillRate,
	}

	go rl.cleanup()

	return rl
}

// isteği kontrol et
func (rl *RateLimiter) Allow(identifier string, n int) bool {
	bucket := rl.getBucket(identifier)

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	bucket.refill()

	if bucket.tokens >= n {
		bucket.tokens -= n
		return true
	}

	return false
}

func (rl *RateLimiter) getBucket(identifier string) *TokenBucket {
	rl.mu.RLock()
	bucket, exists := rl.buckets[identifier]
	rl.mu.RUnlock()

	if exists {
		return bucket
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if bucket, exists := rl.buckets[identifier]; exists {
		return bucket
	}

	bucket = NewTokenBucket(rl.defaultCapacity, rl.defaultRefillRate)
	rl.buckets[identifier] = bucket
	return bucket
}

// identifier için özel limit
func (rl *RateLimiter) SetCustomLimit(identifier string, capacity, refillRate int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.buckets[identifier] = NewTokenBucket(capacity, refillRate)
}

// Bucket'i silme
func (rl *RateLimiter) Remove(identifier string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.buckets, identifier)
}

// kullanılmayan bucketleri silme
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()

		now := time.Now()
		for identifier, bucket := range rl.buckets {
			if now.Sub(bucket.lastRefill) > 10*time.Minute {
				delete(rl.buckets, identifier)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Stats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"active_buckets":      len(rl.buckets),
		"default_capacity":    rl.defaultCapacity,
		"default_refill_rate": rl.defaultRefillRate,
	}
}
