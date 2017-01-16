# otunnel 用法

前提：

1. 假设 server 的地址为 example.com
2. client 与 server 可以不在同一个网络，但是 client 要能访问 server

**注意** otunnel 程序可以分饰 server 和 client 两种角色（运行参数不同）


## 快速上手

### server

```
./otunnel listen :10000 -s longlonglongsecret
```

### client

#### 反向代理

举例：将 client 可以访问的 192.168.1.3:22 映射到 server 上的 10022 端口：

```
./otunnel connect example.com:10000 -s longlonglongsecret -t 'r:192.168.1.3:22::10022'
```

现在访问 example.com:10022 即等于访问了 client 内网的 192.168.1.3:22 !

#### 正向代理

举例：假设 example.com 的 127.0.0.1 网络有 3128 端口（你懂得），在 client 执行：

```
./otunnel connect example.com:10000 -s longlonglongsecret -t 'f:127.0.0.1:20080:127.0.0.1:3128'
```

现在 client 上的任何程序访问 20080 等于访问了 example.com 上的本地 3128 端口。

如果希望 client 局域网内其他机器也能访问 20080 ，需要这样执行：

```
./otunnel connect example.com:10000 -s longlonglongsecret -t 'f::20080:127.0.0.1:3128'
```

**注意** 上面命令区别就是本地监听变成所有网口。


## 程序用法

### `-t` 格式

包含多个字段信息，以 `:` 隔开(为空的字段也不能省略`:`)。

```
代理类型:本地地址:本地端口:远程地址:远程端口
```

| 字段    | 含义                       |
|:--------|:--------------------------|
| 代理类型 | r 表示反向代理; f 表示正向代理 |
| 本地地址 | IP或域名                    |
| 本地端口 | 整数                        |
| 远程地址 | IP或域名                    |
| 远程端口 | 整数                        |

**注意**

1. `本地地址` 或 `远程地址` 如果为空，表示所有网口（多用在需要启动 listen server 的时候）
