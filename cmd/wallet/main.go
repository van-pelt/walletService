package main

import (
	"fmt"
	"github.com/van-pelt/walletService/pkg/wallet"
	"github.com/van-pelt/walletTypes/pkg/types"
	"os"
)

type TestAccounts struct {
	Phone   types.Phone
	Balance types.Money
}

func main() {
	svc := &wallet.Service{}
	var accData = []TestAccounts{
		{Phone: "917590330", Balance: 1500},
		{Phone: "917590331", Balance: 320},
		{Phone: "917590332", Balance: 2500},
		{Phone: "917590333", Balance: 100},
		{Phone: "917590334", Balance: 110},
		{Phone: "917590335", Balance: 460},
		{Phone: "917590336", Balance: 1150},
		{Phone: "917590337", Balance: 770},
		{Phone: "917590338", Balance: 328},
		{Phone: "917590339", Balance: 621},
		{Phone: "917590340", Balance: 885},
	}
	for _, accountGen := range accData {
		account, err := svc.RegisterAccount(accountGen.Phone)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = svc.Deposit(account.ID, accountGen.Balance)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	basePath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	filename := basePath + "/dump/account.dump"

	err = svc.ExportToFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = svc.ImportFromFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	svc.PrintAccounts()
}
