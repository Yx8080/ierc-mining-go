package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

var (
	rpc = "https://rpc.ankr.com/eth_goerli"
	// 根据pow挖矿前缀修改
	prefix = "0x00000"
	// 代表交易发起者愿意支付的最大优先费用（tip）。
	//这个费用是用于激励矿工更快地打包交易。
	//如果矿工在指定的块空间内包含了交易，他们将收到这个费用。
	gasTip = 3
	//这个可以根据链上gas做为调整
	gasMax = 35
	tick   = "ierc-test"
	amt    = 1000
)

var (
	priv      *ecdsa.PrivateKey
	address   common.Address
	ethClient *ethclient.Client
	hexData   string
)
var (
	globalNonce  = time.Now().UnixNano()
	zeroAddress  = common.HexToAddress("0x0000000000000000000000000000000000000000")
	chainID      = big.NewInt(0)
	userNonce    = -1
	successCount int64
)

func main() {
	privateKey, count := inputPrvAndCount()
	println(privateKey)
	hexData = fmt.Sprintf(`data:application/json,{"p":"ierc-20","op":"mint","tick":"%s","amt":"%d","nonce":"%%d"}`, tick, amt)
	log.Infoln("ierc-20 pow mining begins...")
	log.Infoln("Mining begins please wait ...")
	log.Infoln(hexData)

	defer submitSuccessfulLog()

	var err error
	ethClient, err = ethclient.Dial(rpc)
	if err != nil {
		panic(err)
	}

	chainID, err = ethClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	bytePriv, err := hexutil.Decode(privateKey)
	if err != nil {
		panic(err)
	}
	prv, _ := btcec.PrivKeyFromBytes(bytePriv)
	priv = prv.ToECDSA()
	address = crypto.PubkeyToAddress(*prv.PubKey().ToECDSA())
	log.WithFields(log.Fields{
		"prefix":   prefix,
		"amt":      amt,
		"tick":     tick,
		"count":    count,
		"address":  address.String(),
		"chain_id": chainID.Int64(),
	}).Info("prepare done")

	startNonce := globalNonce
	go func() {
		for {
			last := globalNonce
			time.Sleep(time.Second * 10)
			log.WithFields(log.Fields{
				"hash_rate":  fmt.Sprintf("%dhashes/s", (globalNonce-last)/10),
				"hash_count": globalNonce - startNonce,
			}).Info()
		}
	}()

	wg := new(sync.WaitGroup)
	for i := 0; i < count; i++ {
		tx := makeBaseTx()
		wg.Add(runtime.NumCPU())
		ctx, cancel := context.WithCancel(context.Background())
		for j := 0; j < runtime.NumCPU(); j++ {
			go func(ctx context.Context, cancelFunc context.CancelFunc) {
				for {
					select {
					case <-ctx.Done():
						wg.Done()
						return
					default:
						makeTx(cancelFunc, tx)
					}
				}
			}(ctx, cancel)
		}
		wg.Wait()
	}
}

func inputPrvAndCount() (privateKey string, count int) {
	fmt.Print("请输入私钥: ")
	_, err := fmt.Scan(&privateKey)
	if err != nil {
		fmt.Println("读取私钥时发生错误:", err)
		return "", 0
	}

	fmt.Print("请输入铸造数量: ")
	_, err = fmt.Scan(&count)
	if err != nil {
		fmt.Println("读取数量时发生错误:", err)
		return "", 0
	}

	return "0x" + privateKey, count
}

func makeTx(cancelFunc context.CancelFunc, innerTx *types.DynamicFeeTx) {
	atomic.AddInt64(&globalNonce, 1)
	temp := fmt.Sprintf(hexData, globalNonce)
	innerTx.Data = []byte(temp)
	tx := types.NewTx(innerTx)
	signedTx, _ := types.SignTx(tx, types.NewCancunSigner(chainID), priv)
	if strings.HasPrefix(signedTx.Hash().String(), prefix) {
		log.WithFields(log.Fields{
			"tx_hash": signedTx.Hash().String(),
			"data":    temp,
		}).Info("found new transaction")

		err := ethClient.SendTransaction(context.Background(), signedTx)
		if err != nil {
			log.WithFields(log.Fields{
				"tx_hash": signedTx.Hash().String(),
				"err":     err,
			}).Error("failed to send transaction")
		} else {
			log.WithFields(log.Fields{
				"tx_hash": signedTx.Hash().String(),
			}).Info("broadcast transaction submitted successfully")
			// 在成功提交交易时增加计数器
			atomic.AddInt64(&successCount, 1)
		}

		cancelFunc()
	}
}

func makeBaseTx() *types.DynamicFeeTx {
	if userNonce < 0 {
		nonce, err := ethClient.PendingNonceAt(context.Background(), address)
		if err != nil {
			panic(err)
		}
		userNonce = int(nonce)
	} else {
		userNonce++
	}
	innerTx := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     uint64(userNonce),
		GasTipCap: new(big.Int).Mul(big.NewInt(1000000000), big.NewInt(int64(gasTip))),
		GasFeeCap: new(big.Int).Mul(big.NewInt(1000000000), big.NewInt(int64(gasMax))),
		Gas:       30000 + uint64(rand.Intn(1000)),
		To:        &zeroAddress,
		Value:     big.NewInt(0),
	}

	return innerTx
}

func submitSuccessfulLog() {
	log.Infof("Total number of successful transactions committed :%d", successCount)
}
