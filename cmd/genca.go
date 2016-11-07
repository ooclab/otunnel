package cmd

import (
	"github.com/urfave/cli"
	"github.com/ooclab/otunnel/utils"
)

var gencaFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "cn",
		Value: "China",
		Usage: "country name",
	},
	cli.StringFlag{
		Name:  "org",
		Value: "OOCLAB",
		Usage: "Organization",
	},
	cli.StringFlag{
		Name:  "unit",
		Value: "cloud",
		Usage: "Organizational Unit",
	},
	cli.IntFlag{
		Name:  "length",
		Value: 2048,
		Usage: "key length",
	},
}

func gencaAction(c *cli.Context) {
	utils.GenCA(c.String("cn"), c.String("org"), c.String("unit"), c.Int("length"))
}
