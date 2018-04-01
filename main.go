package main

import (
	"fmt"
	"os"

	"github.com/ooclab/otunnel/pkg/client"
	"github.com/ooclab/otunnel/pkg/server"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const programVersion = "1.3.0"

var (
	buildstamp = ""
	githash    = ""
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "01/02 15:04:05",
	})
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println("Current Version: ", c.App.Version)
		if buildstamp != "" {
			fmt.Println("     Build Time: ", buildstamp)
		}
		if githash != "" {
			fmt.Println("Git Commit Hash: ", githash)
		}
	}

	app := cli.NewApp()
	app.Name = "otunnel"
	app.Usage = "otunnel is a simple & secure tunnel tool"
	app.Version = programVersion
	app.Commands = []cli.Command{
		client.Command,
		server.Command,
	}
	app.Run(os.Args)
}
