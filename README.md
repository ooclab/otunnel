# otunnel

**简体中文** | [English](README.en.md)

把两台机器的内网端口，像插网线一样直接互通。

otunnel 是一个单文件、低认知成本的双向安全隧道工具。你可以用同一个命令行，快速完成这些事：

- 临时把本地 SSH/HTTP 暴露给公网团队协作。
- 在本地直接访问远程内网服务做排障和调试。
- 一次连接同时建立多条隧道，组合复杂拓扑。

为什么上手快：

- 一个二进制。
- 一个命令行。
- `listen`（服务端）和 `connect`（客户端）同一套语义。

## 快速开始

1. 从 GitHub Releases 下载对应平台二进制。
2. 在公网服务器启动 otunnel 服务端。
3. 在本地机器启动 otunnel 客户端并建立隧道。

### 安装（Linux amd64 示例）

```bash
curl -fL -o otunnel.tar.gz \
  https://github.com/ooclab/otunnel/releases/download/v1.4.0/otunnel_linux_amd64.tar.gz
tar -xzf otunnel.tar.gz
chmod +x otunnel
sudo mv otunnel /usr/local/bin/otunnel
```

### 30 秒上手

服务端（公网机器）:

```bash
otunnel listen :10000 -d -s abc123
```

客户端（本地机器）:

```bash
otunnel connect SERVER_IP:10000 -d -s abc123 -t 'r:127.0.0.1:22::50022'
```

随后即可通过公网服务器 `SERVER_IP:50022` 访问本地 SSH。

### 常见映射玩法

| 场景 | 命令示例 | 结果 |
| --- | --- | --- |
| 本地服务暴露到公网（反向） | `otunnel connect SERVER_IP:10000 -d -s abc123 -t 'r:127.0.0.1:22::50022'` | 外部访问 `SERVER_IP:50022` 即到本地 `22` |
| 远端服务映射到本地（正向） | `otunnel connect SERVER_IP:10000 -d -s abc123 -t 'f:tcp:127.0.0.1:18080:127.0.0.1:8080'` | 本地访问 `127.0.0.1:18080` 即到远端 `8080` |
| 一次连接，多条隧道 | `otunnel connect SERVER_IP:10000 -d -s abc123 -t 'r:127.0.0.1:22::50022' -t 'f:tcp:127.0.0.1:18080:127.0.0.1:8080'` | 同时启用反向和正向映射 |

可读性版（多行写法）：

```bash
otunnel connect SERVER_IP:10000 -d -s abc123 \
  -t 'r:127.0.0.1:22::50022' \
  -t 'f:tcp:127.0.0.1:18080:127.0.0.1:8080'
```

## 下载

Release 页面:

- [https://github.com/ooclab/otunnel/releases](https://github.com/ooclab/otunnel/releases)

资源下载链接格式:

```text
https://github.com/ooclab/otunnel/releases/download/<version>/otunnel_<os>_<arch>.tar.gz
```

示例:

- [https://github.com/ooclab/otunnel/releases/download/v1.4.0/otunnel_linux_amd64.tar.gz](https://github.com/ooclab/otunnel/releases/download/v1.4.0/otunnel_linux_amd64.tar.gz)
- [https://github.com/ooclab/otunnel/releases/download/v1.4.0/otunnel_darwin_arm64.tar.gz](https://github.com/ooclab/otunnel/releases/download/v1.4.0/otunnel_darwin_arm64.tar.gz)
- [https://github.com/ooclab/otunnel/releases/download/v1.4.0/otunnel_windows_amd64.zip](https://github.com/ooclab/otunnel/releases/download/v1.4.0/otunnel_windows_amd64.zip)

## 从源码构建

```bash
git clone https://github.com/ooclab/otunnel.git
cd otunnel
go mod tidy
make
```

一次构建全部支持平台二进制:

```bash
make build-all
```

## CI

GitHub Actions workflow 位于 `.github/workflows/ci.yml`。
Push/PR 时会执行 `go mod tidy`、`go mod vendor`、`go build ./...` 和 `go test ./...`。

## Systemd

安装到系统路径:

```bash
install -m 755 ./otunnel /usr/local/bin/otunnel
```

服务端配置 `/etc/systemd/system/otunnel-listen.service`:

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

客户端配置 `/etc/systemd/system/otunnel-connect.service`:

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

服务端:

```bash
docker run --rm -it --net=host ooclab/otunnel-amd64 /otunnel listen :10000 -d -s abc123
```

客户端:

```bash
docker run --rm -it --net=host ooclab/otunnel-amd64 /otunnel connect SERVER_IP:10000 -d -s abc123 -t 'f:127.0.0.1:10022:HOST_IP:HOST_PORT'
```

## 文档与支持

- Wiki: [https://github.com/ooclab/otunnel/wiki](https://github.com/ooclab/otunnel/wiki)
- Issues: [https://github.com/ooclab/otunnel/issues](https://github.com/ooclab/otunnel/issues)

## 微信公众号

欢迎关注微信公众号：SynorAI。

![微信公众号二维码](docs/assets/wechat_synorai_qr.png)
