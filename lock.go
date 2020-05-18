package hsocks5

import "sync"

// KeyLock type
type KeyLock struct {
	locks *sync.Map
	mu    *sync.Mutex
}

// NewKeyLock instance
func NewKeyLock() *KeyLock {
	return &KeyLock{
		locks: &sync.Map{},
		mu:    &sync.Mutex{},
	}
}

// GetLock by key
func (kl *KeyLock) GetLock(key string) *sync.Mutex {
	kl.mu.Lock()
	defer kl.mu.Unlock()
	if l, exist := kl.locks.Load(key); exist {
		return l.(*sync.Mutex)
	}
	l := &sync.Mutex{}
	kl.locks.Store(key, l)
	return l
}

// Lock by key
func (kl *KeyLock) Lock(key string) {
	kl.GetLock(key).Lock()
}

// Unlock by key
func (kl *KeyLock) Unlock(key string) {
	kl.GetLock(key).Unlock()
}
