package main

import (
	"fmt"
	"github.com/van-pelt/walletService/pkg/wallet"
)

func main() {
	svc := &wallet.Service{}
	err := svc.GeneratedRandomData()
	if err != nil {
		fmt.Println(err)
		return
	}
	svc.PrintAllWallet()
}
