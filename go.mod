module github.com/ooclab/otunnel

go 1.26.2

require (
	github.com/ooclab/es v0.0.0-20180404035728-d0fea3dc68d5
	github.com/sirupsen/logrus v1.9.4
	github.com/urfave/cli v1.22.17
	golang.org/x/crypto v0.50.0
)

replace github.com/ooclab/es => ../es

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
)
