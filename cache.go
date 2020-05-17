package hsocks5

import (
	"log"
	"time"

	"github.com/go-redis/redis"

	"github.com/patrickmn/go-cache"
)

// KVCache type
type KVCache struct {
	rClient  *redis.Client
	memCache *cache.Cache
	Set      func(string, string)
	Get      func(string) (string, bool)
}

// NewKVCache instance
func NewKVCache(redisAddr ...string) (rt *KVCache) {

	timeout := 30 * 24 * time.Hour

	rt = &KVCache{}

	if len(redisAddr) > 0 && len(redisAddr[0]) > 0 {

		client := redis.NewClient(&redis.Options{Addr: redisAddr[0]})
		_, err := client.Ping().Result()
		if err != nil {
			log.Println(err)
		} else {
			rt.rClient = client
			rt.Set = func(k, v string) {
				rt.rClient.Set(k, v, timeout)
			}
			rt.Get = func(k string) (string, bool) {
				v, err := rt.rClient.Get(k).Result()
				return v, err != redis.Nil
			}
		}

	}

	// without redis client
	if rt.rClient == nil {

		rt.memCache = cache.New(timeout, 1*time.Minute)
		rt.Set = func(k, v string) {
			rt.memCache.SetDefault(k, v)
		}
		rt.Get = func(k string) (string, bool) {
			v, e := rt.memCache.Get(k)
			if !e {
				return "", e
			}
			return v.(string), e
		}

	}

	return

}
