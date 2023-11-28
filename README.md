# IERC 多核挖矿程序

https://holesky.etherscan.io/address/0xd19162560690227c8f71a21b76129e1eb05575a9

### 使用方式

1. 修改main.go中的代码配置

```go
// rpc
 var (
    rpc = "https://1rpc.io/eth"
    prefix = "0x00000"
    gasTip = 3
    gasMax = 50
    tick = "ierc-m5"
    amt = 1000
 )
```