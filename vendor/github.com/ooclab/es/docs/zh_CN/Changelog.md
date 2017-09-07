# 多MSG

将Link,iSession,Tunnel等消息分成多个类型的MSG(LinkMSG/LinkOMSG, EMSG)分开操作
的性能测试：

```
Benchmark_LinkInnerSessionSingle-4         30000             52546 ns/op
Benchmark_LinkInnerSessionMulti-4          30000             56512 ns/op
```
