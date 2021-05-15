package wallet

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/van-pelt/walletTypes/pkg/types"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var ErrPhoneIsRegistred = errors.New("Phone alredy registred")
var ErrAmountMustBePositive = errors.New("amount must be greater that 0")
var ErrAccountNotFound = errors.New("Account not found")
var ErrNotEnoughBalance = errors.New("Not enough balance")
var ErrPaymentNotFound = errors.New("Payment not found")
var ErrFavoriteNotFound = errors.New("Favorite not found")
var ErrFavoriteIsIsset = errors.New("Favorite is isset")
var ErrNoDumpDataFile = errors.New("No DUMP file")

const (
	favorite_prefix = "favorite_"
	accType         = "accounts"
	payType         = "payments"
	favType         = "favorites"
)

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
	if targetPayment == nil {
		return nil, ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
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
	if targetPayment == nil {
		return nil, ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
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
		return fmt.Errorf("Create(%v):%w", path, err)
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
			return fmt.Errorf("Write():%w", err)
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
		if len(elem) != 3 {
			return errors.New("unconsistence data")
		}
		accID, err := strconv.ParseInt(elem[0], 10, 64)
		if err != nil {
			fmt.Printf("%d of type %T", accID, accID)
			return err
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

func (s *Service) PrintAllWallet() {
	for _, elem := range s.accounts {
		fmt.Println("ID=", elem.ID, " Phone=", elem.Phone, " Balance=", elem.Balance)
		fmt.Println("	Payments:")
		for _, p := range s.payments {
			if p.AccountID == elem.ID {
				fmt.Println("		ID=", p.ID, " AccountID=", p.AccountID, " Status=", p.Status, " Category=", p.Category, " Amount=", p.Amount)
			}
		}
		fmt.Println("	Favorite:")
		for _, f := range s.favorites {
			if f.AccountID == elem.ID {
				fmt.Println("		ID=", f.ID, " Account=", f.AccountID, " Name=", f.Name, " Category=", f.Category, " Amount=", f.Amount)
			}
		}
	}
}

func (s *Service) GeneratedRandomData() error {
	phones := []types.Phone{
		"917590330",
		"917590331",
		"917590332",
		"917590333",
		"917590334",
		"917590335",
		"917590336",
		"917590337",
		"917590338",
		"917590339",
		"917590340",
		"917590341",
		"917590342",
		"917590343",
		"917590344",
		"917590345",
		"917590346",
		"917590347",
		"917590348",
		"917590349",
		"917590350",
		"917590351",
	}
	categories := []types.PaymentCategory{
		"auto",
		"internet",
		"food",
		"health",
		"learn",
		"game",
	}
	maxPayment := 40
	//maxFavorite := 5
	maxAmount := 10000
	//maxPhone := len(phones)
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	for _, phone := range phones {
		newAccount, err := s.RegisterAccount(phone)
		if err == ErrPhoneIsRegistred {
			fmt.Println(ErrPhoneIsRegistred)
			continue
		}
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = s.Deposit(newAccount.ID, types.Money(randInt(rnd, maxAmount)))
		if err != nil {
			fmt.Println(err)
			return err
		}

		paymentsCount := randInt(rnd, maxPayment)
		var tmpIds []string
		for i := 0; i < paymentsCount; i++ {
			newPayment, err := s.Pay(newAccount.ID, types.Money(randInt(rnd, 1500)), categories[randInt(rnd, len(categories)-1)])
			if err != nil && err != ErrNotEnoughBalance {
				fmt.Println(err)
				return nil
			}
			if err == ErrNotEnoughBalance {
				continue
			}
			tmpIds = append(tmpIds, newPayment.ID)
		}
		if len(tmpIds) != 0 {
			if len(tmpIds) == 1 {
				_, err := s.FavoritePayment(tmpIds[0], "TEST_FAVOR_0")
				if err != nil {
					return nil
				}
			} else {
				for i := 0; i < randInt(rnd, len(tmpIds)); i++ {
					_, err := s.FavoritePayment(tmpIds[randInt(rnd, len(tmpIds)-1)], "TEST_FAVOR_"+strconv.Itoa(i))
					if err != nil {
						return nil
					}
				}
			}
		}
	}
	return nil
}

func randInt(rand *rand.Rand, max int) int {
	num := rand.Intn(max)
	if num == 0 {
		return 1
	}
	return num
}

func (s *Service) Export(dir string) error {
	if err := s.exportAccounts(dir); err != nil {
		return err
	}
	if err := s.exportPayments(dir); err != nil {
		return err
	}
	if err := s.exportFavorites(dir); err != nil {
		return err
	}
	return nil
}

func (s *Service) exportAccounts(dir string) error {
	if err := s.exportData(dir, accType); err != nil {
		return err
	}
	return nil
}

func (s *Service) exportPayments(dir string) error {
	if err := s.exportData(dir, payType); err != nil {
		return err
	}
	return nil
}

func (s *Service) exportFavorites(dir string) error {
	if err := s.exportData(dir, favType); err != nil {
		return err
	}
	return nil
}

func (s *Service) exportData(dir, exportName string) error {
	if len(s.accounts) != 0 {
		file, err := os.Create(dir + "/" + exportName + ".dump")
		defer func() {
			if cerr := file.Close(); cerr != nil {
				log.Print(err)
			}
		}()
		if err != nil {
			return fmt.Errorf("Create %v,err:%w", exportName, err)
		}

		switch exportName {
		case accType:
			{
				p := types.AccToDump(s.accounts)
				_, err = file.Write([]byte(p.ToDump()))
				if err != nil {
					return fmt.Errorf("Account write is err:%w", err)
				}
			}
		case payType:
			{
				p := types.PayToDump(s.payments)
				_, err = file.Write([]byte(p.ToDump()))
				if err != nil {
					return fmt.Errorf("Payment write is err:%w", err)
				}
			}
		case favType:
			{
				p := types.FavToDump(s.favorites)
				_, err = file.Write([]byte(p.ToDump()))
				if err != nil {
					return fmt.Errorf("Favorite write is err:%w", err)
				}
			}
		default:
			{
				return fmt.Errorf("Unknown type struct")
			}
		}
	}
	return nil
}

func (s *Service) Import(dir string) error {

	if err := s.importAccounts(dir + "/" + accType + ".dump"); err != nil {
		return fmt.Errorf("importAccounts:%w", err)
	}
	if err := s.importPayments(dir + "/" + payType + ".dump"); err != nil {
		return fmt.Errorf("importPayments:%w", err)
	}
	if err := s.importFavorites(dir + "/" + favType + ".dump"); err != nil {
		return fmt.Errorf("importFavorites:%w", err)
	}
	return nil
}

func (s *Service) ParseFile(filepath string) ([]string, error) {
	var data []string
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		log.Printf("WARN:%v.%v file is missing from the set. Data integrity may be compromised", ErrNoDumpDataFile, filepath)
		return data, nil
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("ParseFile.Read() err:%w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data = append(data, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ParseFile.scan err:%w", err)
	}

	return data, nil
}

func (s *Service) importAccounts(filepath string) error {

	data, err := s.ParseFile(filepath)
	if err != nil {
		return err
	}

	if len(data) > 0 {
		for _, elem := range data {
			dataLine := strings.Split(elem, ";") //id,phone,balance
			id, err := strconv.ParseInt(dataLine[0], 10, 64)
			if err != nil {
				return fmt.Errorf("ID ParseInt() err:%w", err)
			}
			acc, errAcc := s.FindAccountByID(id)

			if errAcc != nil && errAcc != ErrAccountNotFound {
				return errAcc
			}
			phone := types.Phone(dataLine[1])
			balance, err := strconv.ParseInt(dataLine[2], 10, 64)
			if err != nil {
				return fmt.Errorf("Balance ParseInt() err:%w", err)
			}

			if errAcc == ErrAccountNotFound {
				acc, err = s.RegisterAccount(types.Phone(dataLine[1]))
				if err != nil {
					return fmt.Errorf("RegisterAccount() err:%w", err)
				}
				err = s.Deposit(acc.ID, types.Money(balance))
				if err != nil {
					return fmt.Errorf("Deposit() err:%w", err)
				}
				continue
			}
			acc.Phone = phone
			acc.Balance = types.Money(balance)
		}
	}
	return nil
}

func (s *Service) importPayments(filepath string) error {

	data, err := s.ParseFile(filepath)
	if err != nil {
		return err
	}

	if len(data) > 0 {
		for _, elem := range data {
			dataLine := strings.Split(elem, ";") //id,accId,amount,Category,status
			payment, errPay := s.FindPaymentByID(dataLine[0])
			if errPay != nil && errPay != ErrPaymentNotFound {
				return err
			}
			accID, err := strconv.ParseInt(dataLine[1], 10, 64)
			if err != nil {
				return fmt.Errorf("ID ParseInt() err:%w", err)
			}
			money, err := strconv.ParseInt(dataLine[2], 10, 64)
			if err != nil {
				return fmt.Errorf("ID ParseInt() err:%w", err)
			}
			amount := types.Money(money)
			category := types.PaymentCategory(dataLine[3])

			_, err = s.FindAccountByID(accID)
			if err != nil && err != ErrAccountNotFound {
				return err
			}
			if err == ErrAccountNotFound {
				log.Printf("WARN:Account not found.Data integrity is violated. Add the accounts.dump file to the sequence.")
				continue
			}
			if errPay == ErrPaymentNotFound {
				s.payments = append(s.payments, &types.Payment{
					ID:        dataLine[0],
					AccountID: accID,
					Amount:    amount,
					Category:  category,
					Status:    types.PaymentStatus(dataLine[4]),
				})
				continue
			} else {
				payment.Amount = amount
				payment.Category = category
				payment.Status = types.PaymentStatus(dataLine[4])
			}
		}
	}
	return nil
}

func (s *Service) importFavorites(filepath string) error {

	data, err := s.ParseFile(filepath)
	if err != nil {
		return err
	}
	if len(data) > 0 {
		for _, elem := range data {
			dataLine := strings.Split(elem, ";") //id,accId,name,amount,Category
			accID, err := strconv.ParseInt(dataLine[1], 10, 64)
			if err != nil {
				return fmt.Errorf("ID ParseInt() err:%w", err)
			}
			name := dataLine[2]
			money, err := strconv.ParseInt(dataLine[3], 10, 64)
			if err != nil {
				return fmt.Errorf("ID ParseInt() err:%w", err)
			}
			amount := types.Money(money)
			category := types.PaymentCategory(dataLine[4])
			favor, err := s.FindFavoritePaymentByID(dataLine[0])
			if err != nil && err != ErrFavoriteNotFound {
				return err
			}
			if err == ErrFavoriteNotFound { //create new favorite by paymentID
				for _, payment := range s.payments {
					if payment.AccountID == accID && payment.Amount == amount && payment.Category == category {
						s.favorites = append(s.favorites, &types.Favorite{
							ID:        dataLine[0],
							AccountID: accID,
							Name:      dataLine[2],
							Amount:    amount,
							Category:  category,
						})
						continue
					}
				}
				continue
			} else {
				favor.AccountID = accID
				favor.Amount = amount
				favor.Name = name
				favor.Category = category
			}
		}
	}
	return nil
}

func (s *Service) ExportAccountHistory(accountID int64) ([]types.Payment, error) {
	var data []types.Payment
	_, err := s.FindAccountByID(accountID)
	if err != nil {
		return data, err
	}
	for _, elem := range s.payments {
		if elem.AccountID == accountID {
			data = append(data, *elem)
		}
	}
	return data, nil
}

func (s *Service) HistoryToFile(payments []types.Payment, dir string, records int) error {
	var strToRecord string
	if len(payments) <= records {
		for _, elem := range payments {
			strToRecord += elem.String()
		}
		if err := s.genericFile(dir, 0, strToRecord); err != nil {
			return err
		}
		return nil
	}

	counterFiles := 1
	var counter int
	endFlag := len(payments) % records
	for i := 0; i < len(payments); i++ {
		strToRecord += payments[i].String()
		counter++
		if counter == records {
			if err := s.genericFile(dir, counterFiles, strToRecord); err != nil {
				return err
			}
			counterFiles++
			counter = 0
			strToRecord = ""
		}
	}
	if endFlag != 0 {
		strToRecord = ""
		for _, b := range payments[len(payments)-endFlag:] {
			strToRecord += b.String()
		}
		if err := s.genericFile(dir, counterFiles, strToRecord); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) genericFile(dir string, index int, data string) error {
	var path string
	if index == 0 {
		path = dir + "/payments.dump"
	} else {
		path = dir + "/payments" + strconv.Itoa(index) + ".dump"
	}

	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("Create %v,err:%w", path, err)
	}
	_, err = file.Write([]byte(data))
	if err != nil {
		file.Close()
		return fmt.Errorf("Payments write is err:%w", err)
	}
	return nil
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
