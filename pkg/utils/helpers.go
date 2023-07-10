package utils

import (
	"binanceMS/pkg/db"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func DecryptStrings(encodedCiphertext string) (string, error) {
	keyString := os.Getenv("ENCRYPTION_PASS")

	key := []byte(keyString)
	data, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(data) < 12 {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:12], data[12:]

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func GetAllowedCoins() []string {
	coin := db.CoinPair{}
	coins, err := coin.GetAllCoins()

	allowed_coins := []string{}
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range *coins {
		allowed_coins = append(allowed_coins, strings.Replace(v.Coin, "/", "", 1))
	}

	return allowed_coins
}

type SymbolDetails struct {
	MarkPrice string `json:"markPrice"`
}

type ResultData struct {
	List []SymbolDetails `json:"list"`
}

type IndexPriceReqBybit struct {
	RetCode    int         `json:"retCode"`
	RetMsg     string      `json:"retMsg"`
	Result     ResultData  `json:"result"`
	RetExtInfo interface{} `json:"retExtInfo"`
	Time       int64       `json:"time"`
}

func GetSizeBybit(symbol string, first_order float64) (float64, float64, error) {
	url := "https://api-testnet.bybit.com/v5/market/tickers?category=inverse&symbol=" + symbol

	client := http.Client{}

	res, err := client.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer res.Body.Close()

	price := IndexPriceReqBybit{}
	if err := json.NewDecoder(res.Body).Decode(&price); err != nil {
		return 0, 0, err
	}

	if price.RetCode != 0 {
		return 0, 0, errors.New(price.RetMsg)
	}

	fprice, err := strconv.ParseFloat(price.Result.List[0].MarkPrice, 64)
	if err != nil {
		return 0, 0, err
	}

	return first_order / fprice, fprice, nil
}
