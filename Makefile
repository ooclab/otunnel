# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOTEST=$(GOCMD) test
GODEP=$(GOTEST) -i
GOFMT=gofmt -w
# LDFLAGS=-ldflags "-s"
#LDFLAGS=-ldflags "-s -X main.buildstamp=`date '+%Y-%m-%d_%H:%M:%S_%z'` -X main.githash=`git rev-parse HEAD`"
LDFLAGS=-a -installsuffix cgo -ldflags "-s -X main.buildstamp=`date '+%Y-%m-%d_%H:%M:%S_%z'` -X main.githash=`git rev-parse HEAD`"

PROGRAM_NAME=otunnel

all:
	cd cmd/otunnel; CGO_ENABLED=0 $(GOBUILD) -v $(LDFLAGS) -o $(PROGRAM_NAME)

install:
	$(GOINSTALL) -v

clean:
	@rm $(PROGRAM_NAME)

mac:
	cd cmd/otunnel; CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -v $(LDFLAGS) -o $(PROGRAM_NAME)
linux-64:
	cd cmd/otunnel; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -v $(LDFLAGS) -o $(PROGRAM_NAME)
linux-32:
	cd cmd/otunnel; CGO_ENABLED=0 GOOS=linux GOARCH=386 $(GOBUILD) -v $(LDFLAGS) -o $(PROGRAM_NAME)
windows-64:
	cd cmd/otunnel; CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -v $(LDFLAGS) -o $(PROGRAM_NAME)
windows-32:
	cd cmd/otunnel; CGO_ENABLED=0 GOOS=windows GOARCH=386 $(GOBUILD) -v $(LDFLAGS) -o $(PROGRAM_NAME)
ddwrt:
	cd cmd/otunnel; CGO_ENABLED=0 GOARCH=arm GOOS=linux GOARM=5 $(GOBUILD) -v $(LDFLAGS) -o $(PROGRAM_NAME)
arm:
	cd cmd/otunnel; CGO_ENABLED=0 GOARCH=arm GOOS=linux GOARM=7 $(GOBUILD) -v $(LDFLAGS) -o $(PROGRAM_NAME)
