package wallet

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/van-pelt/walletTypes/pkg/types"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

var ErrPhoneIsRegistred = errors.New("Phone alredy registred")
var ErrAmountMustBePositive = errors.New("amount must be greater that 0")
var ErrAccountNotFound = errors.New("Account not found")
var ErrNotEnoughBalance = errors.New("Not enough balance")
var ErrPaymentNotFound = errors.New("Payment not found")
var ErrFavoriteNotFound = errors.New("Favorite not found")
var ErrFavoriteIsIsset = errors.New("Favorite is isset")

const favorite_prefix = "favorite_"

type ErrCurrPaymentNotFound struct {
	ErrMess string
	ID      string
}

func NewErrCurrPaymentNotFound(errMess string, ID string) *ErrCurrPaymentNotFound {
	return &ErrCurrPaymentNotFound{ErrMess: errMess, ID: ID}
}

func (e ErrCurrPaymentNotFound) Error() string {
	//panic("implement me")
	return e.ErrMess + ":" + e.ID

}

type Service struct {
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
	nextAccountID int64
}

type Error string
type testService struct {
	*Service
}

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
	s.payments = append(s.payments, payment)
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

func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, payment := range s.payments {
		if payment.ID == paymentID {
			return payment, nil
		}
	}
	return nil, ErrPaymentNotFound
	//return nil, NewErrCurrPaymentNotFound(ErrPaymentNotFound.Error(), paymentID)
}
func (s *Service) Reject(paymentID string) error {

	targetPayment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return err
	}
	if targetPayment == nil {
		return ErrPaymentNotFound
	}

	targetAccount, err := s.FindAccountByID(targetPayment.AccountID)
	if err != nil {
		return err
	}
	targetPayment.Status = types.StatusFail
	targetAccount.Balance += targetPayment.Amount
	return nil
}

func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	targetPayment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}
	if targetPayment == nil {
		return nil, ErrPaymentNotFound
	}

	return &types.Payment{
		ID:        uuid.New().String(),
		AccountID: targetPayment.AccountID,
		Amount:    targetPayment.Amount,
		Category:  targetPayment.Category,
		Status:    targetPayment.Status,
	}, nil
}

func (s *Service) FavoritePayment(paymentID, name string) (*types.Favorite, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("Empty name")
	}
	targetPayment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}
	if targetPayment == nil {
		return nil, ErrPaymentNotFound
	}
	newFavoriteID := favorite_prefix + uuid.New().String()
	_, err = s.FindFavoritePaymentByID(newFavoriteID)
	if err != nil && err != ErrFavoriteNotFound {
		return nil, err
	}
	newFavorite := &types.Favorite{
		ID:        newFavoriteID,
		AccountID: targetPayment.AccountID,
		Name:      name,
		Amount:    targetPayment.Amount,
		Category:  targetPayment.Category,
	}
	s.favorites = append(s.favorites, newFavorite)
	return newFavorite, nil
}

func (s *Service) FindFavoritePaymentByID(favoriteID string) (*types.Favorite, error) {
	for _, favorite := range s.favorites {
		if favorite.ID == favoriteID {
			return favorite, nil
		}
	}
	return nil, ErrFavoriteNotFound
}

func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	favorite, err := s.FindFavoritePaymentByID(favoriteID)
	if err != nil {
		return nil, err
	}
	return s.Pay(favorite.AccountID, favorite.Amount, favorite.Category)
}

func (s *Service) ExportToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Print(err)
		}
	}()
	for _, account := range s.accounts {
		_, err = file.Write([]byte(account.ToString()))
		if err != nil {
			log.Print(err)
		}
	}
	return nil
}
func (s *Service) ImportFromFile(path string) error {
	//dat, err := ioutil.ReadFile("filename")
	content := make([]byte, 0)
	buf := make([]byte, 4)
	file, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Print(err)
		}
	}()

	for {
		read, err := file.Read(buf)
		if err == io.EOF {
			content = append(content, buf[:read]...)
			break
		}
		if err != nil {
			return err
		}
		content = append(content, buf[:read]...)
	}
	data := string(content)
	data = strings.TrimSuffix(data, "|")
	lines := strings.Split(data, "|")
	for _, line := range lines {
		elem := strings.Split(line, ";")
		fmt.Println(elem)
		if len(elem) != 3 {
			return errors.New("unconsistence data")
		}

		accID, err := strconv.ParseInt(elem[0], 10, 64)
		if err != nil {
			fmt.Printf("%d of type %T", accID, accID)
		}
		acc, err := s.FindAccountByID(accID)
		if err != nil && err != ErrAccountNotFound {
			return err
		}
		newBalance, err := strconv.ParseInt(elem[2], 10, 64)
		if err != nil {
			fmt.Printf("%d of type %T", accID, accID)
		}
		if err == ErrAccountNotFound {
			newAccount, err := s.RegisterAccount(types.Phone(elem[1]))
			if err != nil {
				return err
			}
			err = s.Deposit(newAccount.ID, types.Money(newBalance))
			if err != nil {
				return err
			}
		} else {
			acc.Balance = types.Money(newBalance)
			acc.Phone = types.Phone(elem[1])
		}
	}
	return nil
}

func (s *Service) PrintAccounts() {
	for _, elem := range s.accounts {
		fmt.Println("ID=", elem.ID, " Phone=", elem.Phone, " Balance=", elem.Balance)
	}
}

// my func
/*func (s *Service) Reject(paymentID string) error {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return fmt.Errorf("s.FindPaymentByID() is ERR: %w", err)
	}
	if payment.Status != types.StatusInProgres {
		return nil
	}
	payment.Status = types.StatusFail
	acc, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return err
	}
	acc.Balance += payment.Amount
	payment.Amount = 0
	return nil
}*/
