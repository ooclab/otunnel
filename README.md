# otunnel

otunnel is a simple safe tunnel for peer-to-peer

## Build

```
make
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

[Wiki](https://github.com/ooclab/otunnel/wiki)
