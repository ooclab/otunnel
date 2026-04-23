# otunnel

[简体中文](README.md) | **English**

Secure reverse tunnel for exposing local services through a public server.

## Quick Start

1. Download a binary from GitHub Releases.
2. Run server on a public machine.
3. Run client from local machine and create a tunnel.

### Install (Linux amd64 example)

```bash
curl -fL -o otunnel.tar.gz \
  https://github.com/ooclab/otunnel/releases/download/v1.4.0/otunnel_linux_amd64.tar.gz
tar -xzf otunnel.tar.gz
chmod +x otunnel
sudo mv otunnel /usr/local/bin/otunnel
```

### Run in 30 Seconds

Server:

```bash
otunnel listen :10000 -d -s abc123
```

Client:

```bash
otunnel connect SERVER_IP:10000 -d -s abc123 -t 'r:127.0.0.1:22::50022'
```

Now access your local SSH from public server `SERVER_IP:50022`.

## Download

Release page:

- [https://github.com/ooclab/otunnel/releases](https://github.com/ooclab/otunnel/releases)

Asset URL format:

```text
https://github.com/ooclab/otunnel/releases/download/<version>/otunnel_<os>_<arch>.tar.gz
```

Examples:

- [https://github.com/ooclab/otunnel/releases/download/v1.4.0/otunnel_linux_amd64.tar.gz](https://github.com/ooclab/otunnel/releases/download/v1.4.0/otunnel_linux_amd64.tar.gz)
- [https://github.com/ooclab/otunnel/releases/download/v1.4.0/otunnel_darwin_arm64.tar.gz](https://github.com/ooclab/otunnel/releases/download/v1.4.0/otunnel_darwin_arm64.tar.gz)
- [https://github.com/ooclab/otunnel/releases/download/v1.4.0/otunnel_windows_amd64.zip](https://github.com/ooclab/otunnel/releases/download/v1.4.0/otunnel_windows_amd64.zip)

## Build from Source

```bash
git clone https://github.com/ooclab/otunnel.git
cd otunnel
go mod tidy
make
```

Build all supported platforms:

```bash
make build-all
```

## Systemd

Install binary:

```bash
install -m 755 ./otunnel /usr/local/bin/otunnel
```

Server service file `/etc/systemd/system/otunnel-listen.service`:

```ini
[Unit]
Description=Otunnel Listen Service
After=network.target

[Service]
Type=simple
Restart=on-failure
ExecStart=/usr/local/bin/otunnel listen :20000 -d -s THE_SECRET

[Install]
WantedBy=multi-user.target
```

Client service file `/etc/systemd/system/otunnel-connect.service`:

```ini
[Unit]
Description=Otunnel Connect Service
After=network.target

[Service]
Type=simple
Restart=on-failure
ExecStart=/usr/local/bin/otunnel connect YOUR_SERVER_IP:20000 -d -s THE_SECRET -t "r:127.0.0.1:22::50022"

[Install]
WantedBy=multi-user.target
```

## Docker

Server:

```bash
docker run --rm -it --net=host ooclab/otunnel-amd64 /otunnel listen :10000 -d -s abc123
```

Client:

```bash
docker run --rm -it --net=host ooclab/otunnel-amd64 /otunnel connect SERVER_IP:10000 -d -s abc123 -t 'f:127.0.0.1:10022:HOST_IP:HOST_PORT'
```

## Documentation and Support

- Wiki: [https://github.com/ooclab/otunnel/wiki](https://github.com/ooclab/otunnel/wiki)
- Issues: [https://github.com/ooclab/otunnel/issues](https://github.com/ooclab/otunnel/issues)
