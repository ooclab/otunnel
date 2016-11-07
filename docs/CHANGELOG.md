# 版本更新说明

## v0.2.4

### Fix Bugs

1. emsg.ServeTCPLink 

   Fix: panic: send on closed channel (旧机制中，有可能quit已经close了还被发送true。新机制使用close广播)
