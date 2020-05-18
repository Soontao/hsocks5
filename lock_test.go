package hsocks5

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewKeyLock(t *testing.T) {
	l := NewKeyLock()
	k := "v1"
	var t1, t2 time.Time
	wg := sync.WaitGroup{}

	wg.Add(1)
	wg.Add(1)

	go func() {
		l.Lock(k)
		time.Sleep(time.Second)
		t1 = time.Now()
		l.Unlock(k)
		wg.Done()
	}()

	go func() {
		l.Lock(k)
		t2 = time.Now()
		l.Unlock(k)
		wg.Done()
	}()

	wg.Wait()

	dur := t1.Sub(t2).Microseconds()

	// with some locks, it must wait some time
	assert.True(t, dur >= time.Second.Microseconds())

}
