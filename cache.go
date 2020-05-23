package hsocks5

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/go-redis/redis"

	"github.com/patrickmn/go-cache"
)

// KVCache type
type KVCache struct {
	rClient  *redis.Client
	memCache *cache.Cache
	Set      func(string, string)
	Get      func(string) (string, bool)
	MustGet  func(string) string
	logger   *log.Entry
}

// NewKVCache instance
func NewKVCache(redisAddr ...string) (rt *KVCache) {

	logger := log.WithField("module", "hsocks5-cache")

	timeout := 30 * 24 * time.Hour

	rt = &KVCache{logger: logger}

	if len(redisAddr) > 0 && len(redisAddr[0]) > 0 {

		client := redis.NewClient(&redis.Options{Addr: redisAddr[0]})
		_, err := client.Ping().Result()
		if err != nil {
			logger.Println(err)
		} else {
			logger.Println("with redis cache server")
			rt.rClient = client
			rt.Set = func(k, v string) {
				rt.rClient.Set(k, v, timeout)
			}
			rt.Get = func(k string) (string, bool) {
				v, err := rt.rClient.Get(k).Result()
				return v, err != redis.Nil
			}
			rt.MustGet = func(k string) string {
				if r, exist := rt.Get(k); exist {
					return r
				}
				return ""
			}
		}

	}

	// without redis client
	if rt.rClient == nil {

		logger.Println("with in-memory cache")

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
		rt.MustGet = func(k string) (r string) {
			r, _ = rt.Get(k)
			return
		}

	}

	return

}
