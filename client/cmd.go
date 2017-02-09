package client

import (
	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

// Command run connect command
var Command = cli.Command{
	Name:  "connect",
	Usage: "connect to a server",
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
			Usage: "secret phrase",
		},
		cli.StringSliceFlag{
			Name:  "t, tunnel",
			Usage: "new tunnel",
		},
	},
	Action: func(c *cli.Context) {
		if c.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}

		_client, err := newClient(c)
		if err != nil {
			logrus.Errorf("connect to server failed: %s", err)
			return
		}
		_client.Start()
	},
}
