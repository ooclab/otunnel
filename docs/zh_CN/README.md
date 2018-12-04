# otunnel 工具



## 目录

1. [设计](DESIGN.md)
2. [也许](Maybe.md)
3. [相关项目](docs/related.md)

### 平台

- [DD-WRT](./ddwrt.md)



## 架构

otunnel 模块说明

### client/peer/node


### relay server

在节点之间 **转发** TCP/UDP 数据包


### tracker

节点通过连接 tracker ，可以尝试获取对方/自己的 UDP 端口信息：

- client A 向 stun 询问自己的端口，再向 tracker 报告自己的连接信息
- client B 向 tracker 询问 client B 的连接信息，并尝试连接
- client A 与 B 交换数据

tracker 通常只用一个唯一串匹配资源。我们可以把每个连接 tracker 的 client
都设置一个 UID 。其他想建立连接的 client 可以使用该 UID 通过 tracker 找到该 client 。


### stun

### DHT
