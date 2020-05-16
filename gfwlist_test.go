package hsocks5

import (
	"testing"

	"github.com/pmezard/adblock/adblock"
	"github.com/stretchr/testify/assert"
)

func TestLoadGFWList(t *testing.T) {

	matcher := LoadGFWList()
	matchTest := func(url string, expected bool) bool {
		result, _, err := matcher.Match(&adblock.Request{
			URL: url,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expected, result, "%s expected %v, received %v", url, expected, result)
		return result
	}

	matchTest("http://qq.com", false)
	matchTest("http://google.com", false)
	matchTest("https://google.com", false)
	matchTest("https://google.com", false)
	matchTest("https://www.google.com", true)
	matchTest("https://twitter.com", true)
	matchTest("https://facebook.com", true)

}
