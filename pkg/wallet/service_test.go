package wallet

import (
	"github.com/van-pelt/walletTypes/pkg/types"
	"testing"
)

type Mock map[int64]types.Phone
type MockPayments struct {
	accountID int64
	amount    types.Money
	category  types.PaymentCategory
}

func TestService_FindAccountByID_success(t *testing.T) {

	var accountId int64 = 4
	svc, err := GeneratedAccounts()
	if err != nil {
		t.Errorf("want 4, got ERR:%v", err)
		t.Fail()
		return
	}
	acc, err := svc.FindAccountByID(accountId)
	if err != nil {
		t.Errorf("want 4, got ERR:%v", err)
		t.Fail()
		return
	}
	if acc.ID != accountId {
		t.Errorf("want 4, got ERR:%v", acc.ID)
		t.Fail()
		return
	}

}

func TestService_FindAccountByID_NotFound(t *testing.T) {

	var accountId int64 = 44
	svc, err := GeneratedAccounts()
	if err != nil {
		t.Errorf("want 4, got ERR:%v", err)
		t.Fail()
		return
	}
	_, err = svc.FindAccountByID(accountId)
	if err != nil {
		switch err {
		case ErrAccountNotFound:
			t.Logf("Error \"%v\" is OK", err)
		default:
			t.Errorf("want 44, got ERR:%v", err)
			t.Fail()
		}
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

	svc, uuids, err := GeneratedPayments()
	if err != nil {
		t.Errorf("GeneratedPayments() got ERR:%v", err)
		t.Fail()
	}

	for i, v := range uuids {
		payment, err := svc.FindPaymentByID(v)
		if err != nil {
			t.Errorf("svc.FindPaymentByID(%v)->got ERR:%v,iter=%v", v, err, i)
			t.Fail()
		}
		if payment.ID != v {
			t.Errorf("want %v, got :%v", payment.ID, v)
			t.Fail()
		}
	}
}

func TestService_FindPaymentByID_notFound(t *testing.T) {

	svc, uuids, err := GeneratedPayments()
	if err != nil {
		t.Errorf("GeneratedPayments() got ERR:%v", err)
		t.Fail()
	}

	uuids[3] = "46e49cb9-ae24-44ad-bf99-blahblahblah"

	for _, v := range uuids {
		_, err := svc.FindPaymentByID(v)
		if err != nil {
			switch err {
			case ErrPaymentNotFound:
				t.Logf("Error \"%v\" is OK", err)
			default:
				t.Errorf("want %v, got ERR:%v", uuids[3], err)
				t.Fail()
			}
		}
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
			switch err {
			case ErrPaymentNotFound:
				t.Logf("Error \"%v\" is OK", err)
			default:
				t.Errorf("want %v, got ERR:%v", uuids[3], err)
				t.Fail()
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
			switch err {
			case ErrPaymentNotFound:
				t.Logf("Error \"%v\" is OK", err)
			default:
				t.Errorf("want %v, got ERR:%v", v, err)
				t.Fail()
			}
		}
		t.Logf("acc balance after reject=%v", acc.Balance)
	}
}
