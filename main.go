package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	URL = "https://ftx.com/api"
)

var (
	SUB_ACCOUNT   = os.Getenv("SUB_ACCOUNT")
	API_KEY       = os.Getenv("API_KEY")
	SECRET_KEY    = os.Getenv("SECRET_KEY")
	Currency      = os.Getenv("currency")
	UPDATE_MINUTE = os.Getenv("UPDATE_MINUTE")
)

func init() {
	if SUB_ACCOUNT == "" || API_KEY == "" || SECRET_KEY == "" || Currency == "" || UPDATE_MINUTE == "" {
		log.Fatal("plz set .env file")
	}
	log.Printf("Lending Currency is: %s", Currency)
}

func main() {
	minute, _ := strconv.Atoi(UPDATE_MINUTE)
	for {
		balance := GetBalance()
		apy := GetLendingRates()
		SubmitLending(apy, balance)
		time.Sleep(time.Duration(minute) * time.Minute)
	}
}

type Balance struct {
	Success bool `json:"success"`
	Result  []struct {
		Coin  string  `json:"coin"`
		Free  float64 `json:"free"`
		Total float64 `json:"total"`
	} `json:"result"`
}

type LendingRate struct {
	Result []struct {
		Coin     string  `json:"coin"`
		Estimate float64 `json:"estimate"`
		Previous float64 `json:"previous"`
	} `json:"result"`
	Success bool `json:"success"`
}

type LendingOffer struct {
	Coin string  `json:"coin"`
	Size float64 `json:"size"`
	Rate float64 `json:"rate"`
}

type LendingResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func FtxClient(path string, method string, body []byte) *http.Request {
	params := fmt.Sprintf("%s%s/api%s%s", milliTimestamp(), method, path, string(body))
	h := hmac.New(sha256.New, []byte(SECRET_KEY))
	h.Write([]byte(params))
	sign := hex.EncodeToString(h.Sum(nil))
	req, err := http.NewRequest(method, URL+path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("FTX-KEY", API_KEY)
	req.Header.Set("FTX-SIGN", sign)
	req.Header.Set("FTX-TS", milliTimestamp())
	req.Header.Set("FTX-SUBACCOUNT", SUB_ACCOUNT)
	if err != nil {
		return nil
	}
	return req
}

func GetBalance() (totalBalance float64) {
	client := http.Client{}
	path := "/wallet/balances"
	req := FtxClient(path, "GET", nil)
	defer req.Body.Close()
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	r, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println(string(r))
	var balance Balance
	json.Unmarshal(r, &balance)
	for _, coin := range balance.Result {
		if coin.Coin == Currency {
			totalBalance = coin.Total
		}
	}
	return totalBalance
}

func GetLendingRates() (currencyRate float64) {
	client := http.Client{}
	path := "/spot_margin/lending_rates"
	req := FtxClient(path, "GET", nil)
	defer req.Body.Close()
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	r, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var lend LendingRate
	json.Unmarshal(r, &lend)
	for _, coin := range lend.Result {
		if coin.Coin == Currency {
			currencyRate = coin.Estimate * 0.95
		}
	}
	return currencyRate
}

func SubmitLending(apy, balance float64) {
	body, _ := json.Marshal(LendingOffer{Coin: Currency, Size: balance, Rate: apy})
	client := http.Client{}
	path := "/spot_margin/offers"
	req := FtxClient(path, "POST", body)
	defer req.Body.Close()
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	r, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%+v", string(r))

	var lendResp LendingResponse
	json.Unmarshal(r, &lendResp)
	if lendResp.Success == false {
		log.Fatalf("submit lending failed,error: %s", lendResp.Error)
	}

	log.Printf("submit lending success, Currency: %s, Size: %f, APY: %f%%", Currency, balance, apy*24*365*100)
}

func milliTimestamp() string {
	return strconv.FormatInt(time.Now().UTC().Unix()*1000, 10)
}
