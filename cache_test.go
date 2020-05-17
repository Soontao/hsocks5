package hsocks5

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewKVCache(t *testing.T) {

	c := NewKVCache()
	c.Set("a", "1")
	assert.Equal(t, "1", c.MustGet("a"))
	assert.Equal(t, "", c.MustGet("b"))

}

func TestNewKVCacheRedis(t *testing.T) {

	rServer, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	c := NewKVCache(rServer.Addr())
	c.Set("a", "1")
	assert.Equal(t, "1", c.MustGet("a"))
	assert.Equal(t, "", c.MustGet("b"))

}
