package main

import (
	"log"

	"github.com/urfave/cli"
)

var commandStart = cli.Command{
	Name:  "start",
	Usage: "program entry",
	Action: func(c *cli.Context) error {
		log.Println("hello golang.")
		return nil
	},
}
