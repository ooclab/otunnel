# DD-WRT


我使用的硬件是 `Netgear WNDR4300` (Firmware: DD-WRT v3.0-r37882 std (11/30/18))

copy 出路由里的 `/bin/busybox` ，使用 file 查看得知：
```
➜  otunnel git:(master) ✗ file busybox
busybox: ELF 32-bit MSB executable, MIPS, MIPS32 rel2 version 1 (SYSV), dynamically linked, interpreter /lib/ld-musl-mips-sf.so.1, corrupted section header size
```

`MSB` 表示需要使用大端，因此编译参数如下：

```
GOOS=linux GOARCH=mips GOMIPS=softfloat go build -a -ldflags '-s -w'
```

## 参考

- [go-makefile-dd-wrt](https://github.com/lostinblue/go-makefile-dd-wrt) 注意：我仅参考使用了里面的 `GOMIPS=softfloat` 编译项，其他未测试
- [Go 语言跨平台路由器编译](https://blog.lutty.me/code/2017-04/golang-cross-compile-openwrt-ddwrt-build.html)
