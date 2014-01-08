package curr_oracle

import (
	"fmt"
	"net/http"
	"net/url"
	"io/ioutil"
	"encoding/json"
)

const (
	FAKE_GENERATE_ACCOUNT_URL = "http://api-sandbox.oanda.com/v1/accounts"
	ACCOUNT_INFO = "http://api-sandbox.oanda.com/v1/accounts/"
	ACCOUNT_ID = "7476402"
	USER_NAME = "aregory"
	USER_PASS = "Vobnivmow4"
	PLACE_ORDER_URL = "http://api-sandbox.oanda.com/v1/accounts/%d/orders"
)

type accountStruct struct {
	AccountId int `json:"accountId"`
	AccountName string `json:"accountName"`
	Balance float64 `json:"balance"`
	UnrealizedPl float64 `json:"unrealizedPl"`
	RealizedPl float64 `json:"realizedPl"`
	MarginUsed float64 `json:"marginUsed"`
	MarginAvail float64 `json:"marginAvail"`
	OpenTrades float64 `json:"openTrades"`
	OpenOrders float64 `json:"openOrders"`
	MarginRate float64 `json:"marginRate"`
	AccountCurrency string `json:"accountCurrency"`
	Pass string
}


type orderInfo struct {
	Id int64 `json:"id"`
}
type orderStruct struct {
	Time string `json:"time"`
	Price float64 `json:"price"`
	Info *orderInfo `json:"tradeOpened"`
}

func (acc *accountStruct) placeOrder(instrument string, units float64, side string) (orderInfo *orderStruct, err error) {
	resp, err := http.PostForm(fmt.Sprintf(PLACE_ORDER_URL, acc.AccountId),
		url.Values{
			"instrument": {instrument},
			"units": {fmt.Sprintf("%d", int(units))},
			"side": {side},
			"type": {"market"},
		})

	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &orderInfo)
	if err != nil {
		return
	}

	return
}

func fakeAccount(newAccount bool) (account *accountStruct) {
	var resp *http.Response
	var err error

	pass := USER_PASS
	if newAccount {
		var accInfo map[string]interface{}
		resp, err = http.PostForm(FAKE_GENERATE_ACCOUNT_URL, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = json.Unmarshal(body, &accInfo)
		if err != nil {
			fmt.Println(err)
		}
		resp, err = http.Get(fmt.Sprintf("%s%d", ACCOUNT_INFO, int(accInfo["accountId"].(float64))))
		pass = accInfo["password"].(string)
	} else {
		resp, err = http.Get(ACCOUNT_INFO + ACCOUNT_ID)
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal(body, &account)
	if err != nil {
		fmt.Println(err)
	}
	account.Pass = pass

	return
}
