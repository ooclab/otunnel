package main

import (
	"fmt"
	"os"

	// "net/http"
	// _ "net/http/pprof"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/ooclab/otunnel/cmd"
)

const ProgramVersion = "1.1.0"

var (
	buildstamp = ""
	githash    = ""
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "01/02 15:04:05",
	})
}

func main() {
	// ref: http://blog.ralch.com/tutorial/golang-performance-and-memory-analysis/
	// go http.ListenAndServe(":8080", http.DefaultServeMux)

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println("Current Version: ", c.App.Version)
		fmt.Println("     Build Time: ", buildstamp)
		fmt.Println("Git Commit Hash: ", githash)
	}

	app := cli.NewApp()
	app.Name = "otunnel"
	app.Usage = "otunnel is a simple tunnel setup program"
	app.Version = ProgramVersion
	app.Commands = cmd.Commands
	app.Run(os.Args)
}
