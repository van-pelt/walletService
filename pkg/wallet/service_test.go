package wallet

import (
	"errors"
	"fmt"
	"github.com/van-pelt/walletTypes/pkg/types"
	"testing"
)

type Mock map[int64]types.Phone

func TestService_FindAccountByID_success(t *testing.T) {
	acc, err := accountTesting(t, 4)
	if err != nil {
		t.Errorf("want 4, got ERR:%v", err)
		t.Fail()
		return
	}
	if acc.ID != 4 {
		t.Errorf("want 4, got ERR:%v", acc.ID)
		t.Fail()
		return
	}
}

func TestService_FindAccountByID_NotFound(t *testing.T) {
	_, err := accountTesting(t, 44)
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

func accountTesting(t *testing.T, accountId int64) (*types.Account, error) {

	svc, _ := GeneratedAccounts()
	if svc == nil {
		return nil, errors.New("ERR generated account")
	}
	acc, err := svc.FindAccountByID(accountId)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func GeneratedAccounts() (*Service, Mock) {

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
			fmt.Println("acc test ", v, "-", err)
			return nil, mock
		}
	}
	return svc, mock
}
