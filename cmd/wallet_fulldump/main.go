package main

import (
	"fmt"
	"github.com/van-pelt/walletService/pkg/wallet"
)

func main() {
	svc := &wallet.Service{}
	/*err := svc.GeneratedRandomData()
	if err != nil {
		fmt.Println(err)
		return
	}
	//svc.PrintAllWallet()

	/*err = svc.Export("../../dump")
	if err != nil {
		fmt.Println(err)
		return
	}*/

	err := svc.Import("../../dump")
	if err != nil {
		fmt.Println(err)
		return
	}
	svc.PrintAllWallet()
	data, err := svc.ExportAccountHistory(22)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = svc.HistoryToFile(data, "../../dumpPayments", 3)
	if err != nil {
		fmt.Println(err)
		return
	}
}
