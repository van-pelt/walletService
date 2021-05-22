package main

import (
	"fmt"
	"github.com/van-pelt/walletService/pkg/wallet"
	"github.com/van-pelt/walletTypes/pkg/types"
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

	sum := svc.SumPayments(5)
	fmt.Println("current sum=", sum)

	pg := svc.SumPaymentsWithProgress(10)
	sum = 0
	for i := range pg {
		fmt.Println("current Part=", i.Part, " sum=", i.Result)
		sum += i.Result
	}
	fmt.Println("total sum=", sum)

	cnt := svc.FilterPayments(20, 5)
	for _, el := range cnt {
		fmt.Println(el.ID)
	}
	cntf := svc.FilterPaymentsByFN(checkCategory, 5)
	for _, el := range cntf {
		fmt.Println(el.ID, " category=", el.Category, " accountID=", el.AccountID)
	}
}

func checkCategory(payment types.Payment) bool {
	if payment.Category == "learn" && payment.AccountID == 20 {
		return true
	}
	return false
}
