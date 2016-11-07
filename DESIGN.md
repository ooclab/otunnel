otunnel 设计文档

# 原则

Keep It Simple and Stupid

## 方案一

1. otunnel 单程序通过 subcommand 承担 server, client, admin tools 等所有角色。
2. otunnel server 只监听一个指定端口，所有管理、转发等操作都使用一条 connection !
3. otunnel client 默认只是连接上 server 即可。可以通过 otunnel admin tools
   在 client 或 server 端管理隧道（查看、删除、增加、...）
4. otunnel 的 connections 使用自定义协议实现 keepalive (ping/pong)
5. 跨平台：Windows, Mac OS, Linux (跨CPU架构： x86, x86_64, arm)

## 方案二

1. otunnel 所有节点都是 daemon 运行
2. otunnel 使用 http 监听端口,接受管理、连接请求
3. otunnel HTTP daemon 支持如下操作：

   - 启动 client , 连接到指定 daemon
   - 启动 server , 允许其他 otunnel daemon 连接进来
   - 在建立起来的 link 上操作，支持：
     - 创建 tunnel
     - 删除 tunnel
   - 备份配置 (iptables-save)
   - 恢复配置 (iptables-restore)

## 方案三

1. otunnel 只监听1个端口

   - client, server 都需要监听。只是 client 默认只监听本地端口(ip或unix domain socket)，而 server 监听所有端口(ip)
   - client, server 监听端口，都需支持接受管理请求
   - server 的端口还需等待与 client 之间的 connection

2. client, server 有大部分共同的代码
