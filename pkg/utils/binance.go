package utils

import (
	"binanceMS/pkg/db"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

type BinanceLimitResponse struct {
	SymbolStruct []SymbolStruct `json:"symbols"`
}

type SymbolStruct struct {
	Symbol                string   `json:"symbol"`
	Pair                  string   `json:"pair"`
	ContractType          string   `json:"contractType"`
	DeliveryDate          int64    `json:"deliveryDate"`
	OnboardDate           int64    `json:"onboardDate"`
	Status                string   `json:"status"`
	MaintMarginPercent    string   `json:"maintMarginPercent"`
	RequiredMarginPercent string   `json:"requiredMarginPercent"`
	BaseAsset             string   `json:"baseAsset"`
	QuoteAsset            string   `json:"quoteAsset"`
	MarginAsset           string   `json:"marginAsset"`
	PricePrecision        int      `json:"pricePrecision"`
	QuantityPrecision     int      `json:"quantityPrecision"`
	BaseAssetPrecision    int      `json:"baseAssetPrecision"`
	QuotePrecision        int      `json:"quotePrecision"`
	UnderlyingType        string   `json:"underlyingType"`
	UnderlyingSubType     []string `json:"underlyingSubType"`
	SettlePlan            int      `json:"settlePlan"`
	TriggerProtect        string   `json:"triggerProtect"`
	LiquidationFee        string   `json:"liquidationFee"`
	MarketTakeBound       string   `json:"marketTakeBound"`
	MaxMoveOrderLimit     int      `json:"maxMoveOrderLimit"`
	Filters               []Filter `json:"filters"`
	OrderTypes            []string `json:"orderTypes"`
	TimeInForce           []string `json:"timeInForce"`
}

type Filter struct {
	MinPrice   string `json:"minPrice,omitempty"`
	MaxPrice   string `json:"maxPrice,omitempty"`
	FilterType string `json:"filterType"`
	TickSize   string `json:"tickSize,omitempty"`
	StepSize   string `json:"stepSize,omitempty"`
	// MaxQty            string `json:"maxQty,omitempty"`
	// MinQty            string `json:"minQty,omitempty"`
	// Limit             int    `json:"limit,omitempty"`
	// Notional          string `json:"notional,omitempty"`
	// MultiplierDown    string `json:"multiplierDown,omitempty"`
	// MultiplierUp      string `json:"multiplierUp,omitempty"`
	// MultiplierDecimal string `json:"multiplierDecimal,omitempty"`
}

func GetBinanceTradeLimit(coin_pair string) (float64, error) {
	url := "https://www.binance.com/fapi/v1/exchangeInfo?showall=true"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return 0.0, err
	}

	// req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return 0.0, err
	}
	defer res.Body.Close()
	symbols := BinanceLimitResponse{}
	if err := json.NewDecoder(res.Body).Decode(&symbols); err != nil {
		return 0.0, err
	}
	var minPrice string
	for _, symbol := range symbols.SymbolStruct {
		if symbol.Symbol == coin_pair {
			minPrice = symbol.Filters[1].StepSize
			break
		}
	}
	// Check if MinPrice was found
	if minPrice != "" {
		fmt.Println("MinPrice for ", coin_pair, minPrice)
	} else {
		fmt.Println("BTCUSDT not found")
	}
	minPriceFloat, err := strconv.ParseFloat(minPrice, 64)
	if err != nil {
		return 0, err
	}
	return minPriceFloat, nil
}

type IndexPriceReqBinance struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
	Time   int64  `json:"time"`
}

func BinanceRequest(symbol string) (string, error) {
	url := "https://fapi.binance.com/fapi/v1/ticker/price?symbol=" + symbol
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()
	price := IndexPriceReqBinance{}
	if err := json.NewDecoder(res.Body).Decode(&price); err != nil {
		return "", err
	}
	return price.Price, nil
}

func GetBinanceSize(symbol string, first_order float64) (float64, float64, error) {
	binancePrice, err := BinanceRequest(symbol)
	if err != nil {
		return 0, 0, err
	}
	fprice, err := strconv.ParseFloat(binancePrice, 64)
	if err != nil {
		return 0, 0, err
	}
	return first_order / fprice, fprice, nil
}

const (
	BinanceAPIEndpoint = "https://testnet.binancefuture.com"
)

type Position struct {
	Symbol                 string `json:"symbol"`
	InitialMargin          string `json:"initialMargin"`
	MaintMargin            string `json:"maintMargin"`
	UnrealizedProfit       string `json:"unrealizedProfit"`
	PositionInitialMargin  string `json:"positionInitialMargin"`
	OpenOrderInitialMargin string `json:"openOrderInitialMargin"`
	Leverage               string `json:"leverage"`
	Isolated               bool   `json:"isolated"`
	EntryPrice             string `json:"entryPrice"`
	MaxNotional            string `json:"maxNotional"`
	PositionSide           string `json:"positionSide"`
	PositionAmt            string `json:"positionAmt"`
	Notional               string `json:"notional"`
	IsolatedWallet         string `json:"isolatedWallet"`
	UpdateTime             int    `json:"updateTime"`
	BidNotional            string `json:"bidNotional"`
	AskNotional            string `json:"askNotional"`
	LiquidationPrice       string `json:"liquidationPrice"`
	MarkPrice              string `json:"markPrice"`
	MarginType             string `json:"marginType"`
}

type BinanceErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func GenerateBinanceSignature(params map[string]string, secretKey string) string {

	var queryString string
	for key, value := range params {
		queryString += key + "=" + url.QueryEscape(value) + "&"
	}

	queryString = strings.TrimSuffix(queryString, "&")
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(queryString))
	signature := hex.EncodeToString(mac.Sum(nil))

	return signature
}

func GetBinanceAccountOpenPositions(apiKey string, secret string, coinPair string) (*Position, error) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	params := map[string]string{
		"timestamp": strconv.FormatInt(timestamp, 10),
	}

	accSignature := GenerateBinanceSignature(params, secret)
	finalURL := BinanceAPIEndpoint + "/fapi/v2/positionRisk?timestamp=" + strconv.FormatInt(timestamp, 10) + "&signature=" + accSignature

	req, err := http.NewRequest("GET", finalURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-MBX-APIKEY", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse BinanceErrorResponse
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(errorResponse.Msg)
	}

	var positions []Position

	err = json.Unmarshal(body, &positions)
	if err != nil {
		return nil, err
	}

	var requiredPosition Position
	for _, pos := range positions {

		if pos.Symbol == coinPair {
			requiredPosition = pos
			break
		}
	}

	return &requiredPosition, nil
}

func FetchAndUpdateBinancePosition(order *futures.CreateOrderResponse, v db.Key, api_key string, secret_key string, positionSide futures.PositionSideType) {
	positionsResponse, err := GetBinanceAccountOpenPositions(api_key, secret_key, order.Symbol)
	if err != nil {
		fmt.Println("---error getting position data ---", err)
	}

	side := strings.ToLower(string(positionSide))
	oid := fmt.Sprintf("%d", order.OrderID)
	userPosition := db.Positions{
		Symbol:         order.Symbol,
		Leverage:       positionsResponse.Leverage,
		OpenPrice:      positionsResponse.EntryPrice,
		LiqPrice:       positionsResponse.LiquidationPrice,
		TakeProfit:     "0",
		StopLoss:       "0",
		UnrealizedPl:   positionsResponse.UnrealizedProfit,
		MarkPrice:      positionsResponse.MarkPrice,
		Side:           side,
		Size:           order.OrigQuantity,
		Margin:         positionsResponse.InitialMargin,
		UserEmail:      v.UserEmail,
		Status:         "opened",
		Exchange:       "binance",
		OrderId:        oid,
		Layer:          1,
		FirstBuyAmount: positionsResponse.EntryPrice,
	}

	posResponse, createErr := userPosition.CreateNewPosition()
	if createErr != nil {
		fmt.Println(userPosition, "---- error creating new position in database ----", createErr)
		return
	}

	fmt.Println("---- position saved successfully---", posResponse)
}
