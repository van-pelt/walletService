package wallet

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/van-pelt/walletTypes/pkg/types"
	"reflect"
	"testing"
)

type Mock map[int64]types.Phone
type MockPayments struct {
	accountID int64
	amount    types.Money
	category  types.PaymentCategory
}

type TestAccount struct {
	phone    types.Phone
	balance  types.Money
	payments []struct {
		amount   types.Money
		category types.PaymentCategory
	}
}

var defaultTestAccount = TestAccount{
	phone:   "917590333",
	balance: 10000,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{
		{amount: 1000, category: "auto"},
	},
}

func NewTestService() *testService {
	return &testService{&Service{}}
}

func (s *testService) addAccountWithBalance(phone types.Phone, balance types.Money) (*types.Account, error) {
	account, err := s.RegisterAccount(phone)
	if err != nil {
		return nil, fmt.Errorf("Can`t register account,error=%v", err)
	}
	err = s.Deposit(account.ID, balance)
	if err != nil {
		return nil, fmt.Errorf("Can`t deposit account,error=%v", err)
	}
	return account, nil
}

func (s *testService) addAccount(data TestAccount) (*types.Account, []*types.Payment, error) {
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, fmt.Errorf("Can`t register account,error=%v", err)
	}

	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, fmt.Errorf("Can`t deposit account,error=%v", err)
	}
	payments := make([]*types.Payment, len(data.payments))
	for i, payment := range data.payments {
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, fmt.Errorf("Can`t make payment,error=%v", err)
		}
	}
	return account, payments, nil
}

func TestService_FindAccountByID_success(t *testing.T) {
	s := NewTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	payment := payments[0]
	got, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID() error=%v", err)
		return
	}
	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentByID() wrong payment returned %v", err)
		return
	}
}

func TestService_FindAccountByID_NotFound(t *testing.T) {
	s := NewTestService()
	_, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = s.FindPaymentByID(uuid.New().String())
	if err == nil {
		t.Errorf("FindPaymentByID() must return error,returned nil")
		return
	}
	if err != ErrPaymentNotFound {
		t.Errorf("FindPaymentByID() must return ErrPaymentNotFound,returned %v", err)
		return
	}
}

func GeneratedAccounts() (*Service, error) {

	mock := Mock{
		1: "917590333",
		2: "917590334",
		3: "917590335",
		4: "917590336",
		5: "917590337",
		6: "917590338",
	}
	svc := &Service{}
	for _, v := range mock {
		_, err := svc.RegisterAccount(v)
		if err != nil {
			return nil, err
		}
	}
	return svc, nil
}

func GeneratedPayments() (*Service, []string, error) {

	svc := &Service{}
	acc, err := svc.RegisterAccount("917590333")
	if err != nil {
		return svc, nil, err
	}
	err = svc.Deposit(acc.ID, 1500)
	if err != nil {
		return svc, nil, err
	}
	mock := []MockPayments{
		{
			accountID: acc.ID,
			amount:    250,
			category:  "internet",
		},
		{
			accountID: acc.ID,
			amount:    50,
			category:  "phone",
		},
		{
			accountID: acc.ID,
			amount:    150,
			category:  "health",
		},
		{
			accountID: acc.ID,
			amount:    50,
			category:  "internet",
		},
		{
			accountID: acc.ID,
			amount:    40,
			category:  "food",
		},
	}
	var uuids []string
	for _, v := range mock {
		payment, err := svc.Pay(v.accountID, v.amount, v.category)
		if err != nil {
			return svc, nil, err
		}
		uuids = append(uuids, payment.ID)
	}
	return svc, uuids, nil
}

func TestService_FindPaymentByID_success(t *testing.T) {
	s := NewTestService()
	account, err := s.addAccountWithBalance("917590333", 10000)
	if err != nil {
		t.Error(err)
		return
	}
	payment, err := s.Pay(account.ID, 1000, "auto")
	if err != nil {
		t.Errorf("FindPaymentByID() can`t create payment,err=%v", err)
		return
	}
	got, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID() err=%v", err)
		return
	}

	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentByID() wrong payment returned err=%v", err)
		return
	}
}

func TestService_FindPaymentByID_notFound(t *testing.T) {
	s := NewTestService()
	account, err := s.addAccountWithBalance("917590333", 10000)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = s.Pay(account.ID, 1000, "auto")
	if err != nil {
		t.Errorf("FindPaymentByID() can`t create payment,err=%v", err)
		return
	}
	_, err = s.FindPaymentByID(uuid.New().String())
	if err == nil {
		t.Errorf("FindPaymentByID() must return error,returned nil")
		return
	}

	if err != ErrPaymentNotFound {
		t.Errorf("FindPaymentByID() must return ErrPaymentNotFound,returned=%v", err)
		return
	}
}

func TestService_Reject_notFound(t *testing.T) {
	svc, uuids, err := GeneratedPayments()
	if err != nil {
		t.Errorf("GeneratedPayments() got ERR:%v", err)
		t.Fail()
	}

	uuids[3] = "46e49cb9-ae24-44ad-bf99-blahblahblah"

	for _, v := range uuids {
		err := svc.Reject(v)
		if err != nil {
			var newErr *ErrCurrPaymentNotFound
			if errors.As(err, &newErr) {
				t.Logf("Error \"%v\" is OK", newErr.ErrMess)
			}
		}
	}
}

func TestService_Reject_success(t *testing.T) {
	svc, uuids, err := GeneratedPayments()
	if err != nil {
		t.Errorf("GeneratedPayments() got ERR:%v", err)
		t.Fail()
	}

	for _, v := range uuids {
		p, err := svc.FindPaymentByID(v)
		if err != nil {
			t.Errorf("want %v, got ERR:%v", v, err)
			t.Fail()
		}
		t.Logf("Payment %v,amount=%v,status=%v", p.ID, p.Amount, p.Status)
		acc, err := svc.FindAccountByID(p.AccountID)
		if err != nil {
			t.Errorf("want %v, got ERR:%v", v, err)
			t.Fail()
		}
		t.Logf("acc balance=%v", acc.Balance)
		err = svc.Reject(v)
		if err != nil {
			var newErr *ErrCurrPaymentNotFound
			if errors.As(err, newErr) {
				t.Logf("Error \"%v\" is OK", newErr.ErrMess)
			}

			/*if errors.Is(err, ErrPaymentNotFound) {
				t.Logf("Error \"%v\" is OK", err)
			}
			/*switch err {
			case ErrPaymentNotFound:
				t.Logf("Error \"%v\" is OK", err)
			default:
				t.Errorf("want %v, got ERR:%v", v, err)
				t.Fail()
			}*/
		}
		t.Logf("acc balance after reject=%v", acc.Balance)
	}
}

func TestService_Reject_success_lesson(t *testing.T) {
	s := NewTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	payment := payments[0]
	fmt.Println(payment.Status)
	err = s.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject() err=%v", err)
		return
	}
	savedPayment, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("Reject() can`t find payment by ID,err=%v", err)
		return
	}
	fmt.Println(savedPayment.Status)
	if savedPayment.Status != types.StatusFail {
		t.Errorf("Reject() status didn`t changed,payment=%v", savedPayment)
		return
	}
	savedAccount, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		t.Errorf("Reject() can`t find account by ID,err=%v", err)
		return
	}
	if savedAccount.Balance != defaultTestAccount.balance {
		t.Errorf("Reject() balance didn`t changed ,account=%v", savedAccount)
		return
	}
}

func TestService_Repeat_success(t *testing.T) {
	s := NewTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	newPayment, err := s.Repeat(payments[0].ID)
	if err != nil {
		t.Errorf("Repeat() returned err=%v", err)
		return
	}
	if newPayment.ID == payments[0].ID {
		t.Errorf("Repeat() IDs is identical newPaymentID=%v, paymentId=%v", newPayment.ID, payments[0].ID)
		return
	}
	t.Log(newPayment.ID, " ", payments[0].ID)
	t.Log(newPayment, " ", payments[0])
}

func TestService_Repeat_fail(t *testing.T) {
	s := NewTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	newPayment, err := s.Repeat(payments[0].ID)
	payments[0].ID = newPayment.ID
	if err != nil {
		t.Errorf("Repeat() returned err=%v", err)
		return
	}
	if newPayment.ID == payments[0].ID {
		t.Log(newPayment.ID, " ", payments[0].ID)
		t.Logf("Repeat() IDs is identical newPaymentID=%v, paymentId=%v", newPayment.ID, payments[0].ID)
		return
	}
}

func TestService_FavoritePayment_success(t *testing.T) {
	s := NewTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	favorite, err := s.FavoritePayment(payments[0].ID, "test_favorite")
	if err != nil {
		t.Errorf("FavoritePayment() returned err=%v", err)
		return
	}
	if favorite != nil {
		t.Logf("NEW favorite id=%v", favorite.ID)
		return
	}

}

func TestService_FindFavoritePaymentByID_fail(t *testing.T) {
	s := NewTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = s.FavoritePayment(payments[0].ID, "test_my_favor")
	if err != nil {
		t.Error(err)
		return
	}
	favoriteId := "badID"
	_, err = s.FindFavoritePaymentByID(favoriteId)
	if err == ErrFavoriteNotFound {
		t.Logf("FindFavoritePaymentByID() returned ErrFavoriteNotFound [%v]", err)
		return
	}
}

func TestService_FindFavoritePaymentByID_success(t *testing.T) {
	s := NewTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	favorite, err := s.FavoritePayment(payments[0].ID, "test_my_favor")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = s.FindFavoritePaymentByID(favorite.ID)
	if err == ErrFavoriteNotFound {
		t.Error(err)
		return
	}

}

func TestService_FavoritePayment_fail_name(t *testing.T) {
	s := NewTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	name := ""
	_, err = s.FavoritePayment(payments[0].ID, name)
	if err != nil {
		t.Logf("FavoritePayment() returned is empty name error.err=%v", err)
		return
	}

}

func TestService_PayFromFavorite_success(t *testing.T) {
	s := NewTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	name := "test_favorite"
	favorite, err := s.FavoritePayment(payments[0].ID, name)
	if err != nil {
		t.Error(err)
		return
	}
	payment, err := s.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("PayFromFavorite() success.New paymentID=%v", payment.ID)
	return
}

func TestService_PayFromFavorite_fail(t *testing.T) {
	s := NewTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	name := "test_favorite_fail"
	favorite, err := s.FavoritePayment(payments[0].ID, name)
	if err != nil {
		t.Error(err)
		return
	}
	s.favorites[0].Amount = 0
	_, err = s.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Logf("PayFromFavorite() fail is 0 amount.Err=%v", err)
		return
	}

}
