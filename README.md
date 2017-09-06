# otunnel

otunnel is a simple safe tunnel for peer-to-peer

## Build

```
make
```

OR

```
./build-by-docker.sh
```


### use `go build`

```
cd cmd/otunnel
```

```
go build -v
```

build a static bin:

```
$ CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -v
$ ldd otunnel 
        not a dynamic executable
```


### use gox

```
cd cmd/otunnel
```

```
gox
```

OR

```
gox -ldflags "-s -X main.buildstamp=`date '+%Y-%m-%d_%H:%M:%S_%z'` -X main.githash=`git rev-parse HEAD`" 
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

## Document

[Wiki/手册](https://github.com/ooclab/otunnel/wiki)

## Download

[Download](http://dl.ooclab.com/otunnel/)

For example:

```
wget http://dl.ooclab.com/otunnel/1.1.0/otunnel_linux_amd64.xz
unxz otunnel_linux_amd64.xz
chmod a+x otunnel_linux_amd64
mv otunnel_linux_amd64 otunnel
```

## Help

Please send issues to [github.com/ooclab/otunnel/issues](https://github.com/ooclab/otunnel/issues) .

