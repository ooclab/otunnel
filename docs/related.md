# 相关项目

## 后起

多数使用 golang 语言开发，这个现象值得思考。

### [qtunnel](https://github.com/getqujing/qtunnel)

**优点**：简洁、稳定

**缺点**：功能也是如此

我们借鉴了 qtunnel AES 加密方法

### [frp](https://github.com/fatedier/frp/blob/master/README_zh.md)

开发初衷及与 dog-tunnel 区别：https://github.com/fatedier/frp/issues/37

**优点**：

- 目标明确：反向代理
- 更多的功能正在TODO
- 开源

**缺点**：还不够简单

### [ngrok](https://ngrok.com/)

**优点**：功能丰富，商业，有公共服务

**缺点**：看看 https://ngrok.com/docs ，你知道他在说什么吗？我是不知道。

### [dog tunnel](https://github.com/vzex/dog-tunnel)

主页： http://dog-tunnel.tk/

优点：本土化

缺点：...

### 其他

- https://mdn.mengxiaozhu.cn/


## 先见

多数使用 C 开发 :-)

### ssh

正反向代理都支持，但是不太稳定: 不能断开重连。

另外所有基于 TLS/SSL 加密的协议都有个问题，握手是明文的，虽然“安全”，但是不能抗干扰。（如 openvpn）

### openvpn

如果购买一个 openvpn 服务，客户端设置还算简单。
如果自己搭建一个 openvpn 服务，那就费事了。

openvpn 的问题上面提到，不能抗干扰

### socat

功能强大，非 Unix/Linux 高手用不了

### netcat / nc

强大，但有点令人抓狂了。
