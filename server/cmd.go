package server

import (
	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

// Command run listen command
var Command = cli.Command{
	Name:  "listen",
	Usage: "Listen as a server, wait connects from clients.",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "d, debug",
			Usage: "debug log level",
		},
		cli.StringFlag{
			Name:  "P, proto",
			Value: "tcp",
			Usage: "the proto between two points",
		},
		cli.StringFlag{
			Name:  "s, secret",
			Value: "",
			Usage: "secret phrase",
		},
		cli.IntFlag{
			Name:  "keepalive",
			Value: 30,
			Usage: "keepalive interval",
		},
	},
	Action: func(c *cli.Context) {
		if c.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}

		_server := newServer(c)
		_server.Start()
	},
}
