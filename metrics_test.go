package hsocks5

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProxyServerMetrics(t *testing.T) {
	assert.NotNil(t, NewProxyServerMetrics(), "must create metric object")
}
