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

//Yeni bucket oluştur
