setup:

```
go get -v -u github.com/sirupsen/logrus
go get -v -u github.com/jmoiron/jsonq

go get -v -u github.com/ooclab/es
cd $GOPATH/src/github.com/ooclab/es/example
```

build:

```
go build -v server.go
go build -v client.go
```

run server:

```
./server :3000
```

run client example 1 (maping CLIENT:127.0.0.1:8080 to SERVER:*:18080):

```
./client SERVER_IP:3000 r:127.0.0.1:8080::18080
```

run client example 2 (maping SERVER:127.0.0.1:1080 to CLIENT:127.0.0.1:8000):

```
./client SERVER_IP:3000 f:127.0.0.1:8000:127.0.0.1:1080
```
