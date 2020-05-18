package main

import "github.com/urfave/cli"

var options = []cli.Flag{
	&cli.StringFlag{
		Name:   "addr, a",
		EnvVar: "ADDR",
		Usage:  "HTTP proxy listen address",
		Value:  ":18080",
	},
	&cli.StringFlag{
		Name:     "socks, s",
		EnvVar:   "SOCKS",
		Required: true,
		Usage:    "Socks5 service url, format: hostname:port, 192.168.1.1:18080",
	},
	&cli.StringFlag{
		Name:   "redis, r",
		EnvVar: "REDIS_SERVER",
		Usage:  "Redis server for proxy check",
	},
	&cli.BoolFlag{
		Name:   "china-switchy",
		EnvVar: "CHINA_SWITCHY",
		Usage:  "For mainland china user, 'hsocks' can automatic use 'socks5 proxy' ondemand",
	},
}
