# otunnel

[简体中文](README.md) | **English**

Connect two private networks as if they were on the same cable.

otunnel is a single-binary, low-friction, bidirectional secure tunnel CLI. With one tool, you can:

- Share local SSH/HTTP to public teammates in minutes.
- Reach remote private services from local for debugging.
- Run mixed forward/reverse mappings in one connection.

Why it is fast to adopt:

- One binary.
- One CLI.
- Same command model for both roles: `listen` and `connect`.

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

### Common Tunnel Patterns

| Scenario | Command | Result |
| --- | --- | --- |
| Expose local service to public side (reverse) | `otunnel connect SERVER_IP:10000 -d -s abc123 -t 'r:127.0.0.1:22::50022'` | Access `SERVER_IP:50022` to reach local `22` |
| Map remote internal service to local (forward) | `otunnel connect SERVER_IP:10000 -d -s abc123 -t 'f:tcp:127.0.0.1:18080:127.0.0.1:8080'` | Access local `127.0.0.1:18080` to reach remote `8080` |
| Multiple tunnels in one connection | `otunnel connect SERVER_IP:10000 -d -s abc123 -t 'r:127.0.0.1:22::50022' -t 'f:tcp:127.0.0.1:18080:127.0.0.1:8080'` | Reverse + forward tunnels together |

Readable multiline form:

```bash
otunnel connect SERVER_IP:10000 -d -s abc123 \
  -t 'r:127.0.0.1:22::50022' \
  -t 'f:tcp:127.0.0.1:18080:127.0.0.1:8080'
```

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
