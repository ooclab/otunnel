package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ooclab/otunnel/server"
	"github.com/urfave/cli"
)

var listenFlags = []cli.Flag{
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
}

func listenAction(c *cli.Context) {
	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	_server := server.NewServer(c)
	_server.Start()
}

// CommandListen run listen command
var CommandListen = cli.Command{
	Name:   "listen",
	Usage:  "Listen as a server, wait connects from clients.",
	Flags:  listenFlags,
	Action: listenAction,
}
