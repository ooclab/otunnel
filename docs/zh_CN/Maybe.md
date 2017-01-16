# Maybe

## windows 服务

参考： https://github.com/fatedier/frp/issues/35

windows发布，建议增加服务（service 保证windows启动时 s端自动已启动）指导说明：

```
NSSM - the Non-Sucking Service Manager
http://www.nssm.cc/
安装服务： nssm install frps x:\xxx\frp\frps.exe
启动服务：nssm start frps
停止服务：nssm stop frps
```
