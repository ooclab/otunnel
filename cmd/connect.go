package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/ooclab/otunnel/client"
)

var connectFlags = []cli.Flag{
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
	cli.StringFlag{
		Name:  "pprof",
		Value: "",
		Usage: "listen address for pprof",
	},
}

func connectAction(c *cli.Context) {
	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	_client, err := client.NewClient(c)
	if err != nil {
		log.Errorf("connect to server failed: %s", err)
		return
	}
	_client.Start()
}
