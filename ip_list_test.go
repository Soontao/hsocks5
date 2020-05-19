package hsocks5

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrivateIPList(t *testing.T) {
	ipList := LoadIPListFrom("assets/private_ip_list.txt")
	assert.True(t, ipList.Contains("192.168.1.1"))
	assert.True(t, ipList.Contains("10.1.1.1"))
	assert.False(t, ipList.Contains("55.1.1.1"))
}

func TestCNIPList(t *testing.T) {
	ipList := LoadIPListFrom("assets/china_ip_list.txt")
	assert.True(t, ipList.Contains("cn.aliyun.com"))
	assert.True(t, ipList.Contains("111.206.4.92"))    // wx3.sinaimg.cn
	assert.True(t, ipList.Contains("39.156.69.79"))    // baidu.com
	assert.True(t, ipList.Contains("114.114.114.114")) // 114 DNS
	assert.False(t, ipList.Contains("172.217.160.78")) // google
}
