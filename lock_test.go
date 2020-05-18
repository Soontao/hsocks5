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
	t1 := time.Now()
	waitTime := time.Millisecond * 100
	wg := sync.WaitGroup{}

	f := func() {
		l.Lock(k)
		time.Sleep(waitTime)
		l.Unlock(k)
		wg.Done()
	}

	wg.Add(1)
	go f()

	wg.Add(1)
	go f()

	wg.Wait()

	dur := time.Now().Sub(t1).Microseconds()

	expDur := 2 * waitTime.Microseconds()

	// with some locks, it must wait some time
	assert.Truef(t, dur >= expDur, "must spent time more than %v, but spend: %v", expDur, dur)

}
