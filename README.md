# IERC
出处: https://github.com/minchenzz/ierc-miner

https://goerli.etherscan.io/address/0x67F657F66291B350bc9B75EAEF05E52FE229B794
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