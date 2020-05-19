package main

import (
	"github.com/Soontao/hsocks5"
	"github.com/urfave/cli"
)

var commandStart = cli.Command{
	Name:  "start",
	Usage: "program entry",
	Action: func(c *cli.Context) error {
		addr := c.GlobalString("addr")
		socks := c.GlobalString("socks")
		redis := c.GlobalString("redis")
		chinaSwitchy := c.GlobalBool("china-switchy")
		s, err := hsocks5.NewProxyServer(&hsocks5.ProxyServerOption{
			ListenAddr:  addr,
			RedisAddr:   redis,
			ChinaSwitch: chinaSwitchy,
			SocksAddr:   socks,
		})
		if err != nil {
			return err
		}
		return s.Start()
	},
}
