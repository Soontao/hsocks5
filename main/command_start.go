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
		s, err := hsocks5.NewProxyServer(socks)
		if err != nil {
			return err
		}
		s.Start(addr)
		return nil
	},
}
