package main

import "github.com/urfave/cli"

var options = []cli.Flag{
	&cli.Int64Flag{
		Name:   "port, p",
		EnvVar: "PORT",
		Usage:  "HTTP proxy port",
		Value:  18080,
	},
	&cli.StringFlag{
		Name:   "socks, s",
		EnvVar: "SOCKS",
		Usage:  "Socks5 service url",
	},
}
