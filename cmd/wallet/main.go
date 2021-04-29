package main

import (
	"fmt"
	"github.com/van-pelt/walletService/pkg/wallet"
)

func main() {
	svc := &wallet.Service{}
	account, err := svc.RegisterAccount("917590333")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = svc.Deposit(account.ID, 30)
	if err != nil {
		switch err {
		case wallet.ErrAccountNotFound:
			fmt.Println("Пользователь не найден")
		case wallet.ErrAmountMustBePositive:
			fmt.Println("Сумма должна быть положительной")
		}
		return
	}
	fmt.Println(account.Balance)
}
