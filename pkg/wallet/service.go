package wallet

import (
	"errors"
	"github.com/google/uuid"
	"github.com/van-pelt/walletTypes/pkg/types"
)

var ErrPhoneIsRegistred = errors.New("Phone alredy registred")
var ErrAmountMustBePositive = errors.New("amount must be greater that 0")
var ErrAccountNotFound = errors.New("Account not found")
var ErrNotEnoughBalance = errors.New("Not enough balance")

type Service struct {
	accounts      []*types.Account
	payments      []*types.Payment
	nextAccountID int64
}

type Error string

func (e Error) Error() string {
	return string(e)
}

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneIsRegistred
		}
	}
	s.nextAccountID++
	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)
	return account, nil
}

func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}
	account, err := s.FindAccountByID(accountID)
	if err != nil {
		return err
	}

	if account == nil {
		return ErrAccountNotFound
	}
	account.Balance += amount
	return nil
}

func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}
	account, err := s.FindAccountByID(accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}
	if account.Balance < amount {
		return nil, ErrNotEnoughBalance
	}
	account.Balance -= amount
	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.StatusInProgres,
	}
	return payment, nil
}

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			return acc, nil
		}
	}
	return nil, ErrAccountNotFound
}
