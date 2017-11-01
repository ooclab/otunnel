# otunnel

otunnel is a simple safe tunnel for peer-to-peer

## Build

simple build (RECOMMENDED):

```
$ ./build-by-docker.sh
```

others:

```
$ go get -v github.com/ooclab/otunnel
$ export GOPATH=${GOPATH:-~/go}
$ cd $GOPATH/src/github.com/ooclab/otunnel

$ # use any of following commands to build otunnel

$ make                    # normal build
$ make static             # build a static program
$ go build -v             # the go build
$ gox                     # simple cross build, you should install gox first!
```

## Usage

Start a server at a public server ( example.com ):

```
./otunnel listen -d
```

Start a client (reverse forward):

```
./otunnel connect example.com:10000 -d -t r:LOCAL_HOST:LOCAL_PORT::REMOTE_PORT
```

Now, anyone can access your `LOCAL_HOST:LOCAL_PORT` by `example.com:REMOTE_PORT`.

## Docker

Run a server:

```
docker run --rm -it -p 20000:10000 ooclab/otunnel-amd64 /otunnel listen :10000 -d -s abc123
```

Run a client:

```
docker run --rm -it --net=host ooclab/otunnel-amd64 /otunnel connect SERVER_IP:20000 -d -s abc123 -t 'f:127.0.0.1:10022:HOST_IP:HOST_PORT'
```

## Document

[Wiki / 手册](https://github.com/ooclab/otunnel/wiki)

## Download

[Download](http://dl.ooclab.com/otunnel/)

For example:

```
wget http://dl.ooclab.com/otunnel/1.2.3/otunnel_linux_amd64.xz
unxz otunnel_linux_amd64.xz
chmod a+x otunnel_linux_amd64
mv otunnel_linux_amd64 otunnel
```

## Help

Please send issues to [github.com/ooclab/otunnel/issues](https://github.com/ooclab/otunnel/issues) .

## Other Projects

- qtunnel
- ngrok
- frp
- pagekite
