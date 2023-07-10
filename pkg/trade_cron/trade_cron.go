package trade_cron

import (
	"binanceMS/pkg/db"
	"binanceMS/pkg/utils"

	"context"
	"fmt"
	"log"
	"math"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/jinzhu/gorm"
)

type TradeCron struct {
	DB *gorm.DB
}

func NewTradeCron() *TradeCron {
	return &TradeCron{}
}

func (tc *TradeCron) Run() {

	futures.UseTestnet = true

	log.Println("trade cron started")

	key := db.Key{}

	keys, err := key.FindAllKeys()
	if err != nil {
		log.Fatal(err)
	}

	symbol := "APEUSDT" //utils.SelectRandomElement(utils.GetAllowedCoins())
	log.Println(symbol)

	pos := db.Positions{}

	for _, key := range *keys {

		if !key.Start {
			continue
		}

		if key.OpenLong <= 0 || key.OpenShort <= 0 {
			continue
		}

		poss, err := pos.FindAllUserPositions(key.UserEmail)
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range *poss {
			if v.Symbol == symbol {
				log.Println("already position with this symbol")
				return
			}
		}

		val := int(math.Floor(float64(key.TradeAmount)/100.0) * 100)
		cond := db.Conditions{}
		c, err := cond.FindCondition(val)
		if err != nil {
			log.Fatal(err)
		}
		first_order := float64(key.TradeAmount) * 0.08 / float64(c.Positions)
		if key.Mode == "aggressive" {
			first_order = first_order * 2
		}

		x, p, err := utils.GetBinanceSize(symbol, first_order)

		if err != nil {
			log.Fatal(err)
		}

		stepSize, err := utils.GetBinanceTradeLimit(symbol)
		if err != nil {
			println("error in calculating limit", stepSize)
			return
		}
		precision := int(math.Round(-math.Log10(stepSize)))
		quantity := math.Round(x*math.Pow(10, float64(precision))) / math.Pow(10, float64(precision))
		quantity_str := fmt.Sprintf("%f", quantity)
		log.Println(quantity_str)
		api_key, err := utils.DecryptStrings(key.ApiKey)
		if err != nil {
			log.Fatal(err)
		}

		secret_key, err := utils.DecryptStrings(key.SecretKey)
		if err != nil {
			log.Fatal(err)
		}

		BinanceClient := futures.NewClient(api_key, secret_key)

		err = BinanceClient.NewChangePositionModeService().DualSide(true).Do(context.Background())
		if err != nil {
			log.Println(err)
			//	return
		}

		order1, err := BinanceClient.NewCreateOrderService().Symbol(symbol).
			PositionSide(futures.PositionSideTypeLong).
			Side(futures.SideTypeBuy).Type(futures.OrderTypeMarket).
			Quantity(quantity_str).
			Do(context.Background())
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(order1)

		order2, err := BinanceClient.NewCreateOrderService().Symbol(symbol).
			PositionSide(futures.PositionSideTypeShort).
			Side(futures.SideTypeSell).
			Type(futures.OrderTypeMarket).
			Quantity(quantity_str).
			Do(context.Background())
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(order2)

		long_order := db.Order{
			Email:       key.UserEmail,
			Symbol:      order1.Symbol,
			Size:        quantity_str,
			Side:        "open_long",
			MarginCoin:  "USDT",
			OrderType:   "market",
			Service:     "binance",
			QuoteAmount: p,
		}

		short_order := db.Order{
			Email:       key.UserEmail,
			Symbol:      order2.Symbol,
			Size:        quantity_str,
			Side:        "open_short",
			MarginCoin:  "USDT",
			OrderType:   "market",
			Service:     "binance",
			QuoteAmount: p,
		}

		_, err = long_order.SaveOrder()
		if err != nil {
			log.Println(err)
		}
		_, err = short_order.SaveOrder()
		if err != nil {
			log.Println(err)
		}
		key.ChangePositions(key.OpenLong-1, key.OpenShort-1)
		go utils.FetchAndUpdateBinancePosition(order1, key, api_key, secret_key, "long")
		go utils.FetchAndUpdateBinancePosition(order2, key, api_key, secret_key, "short")

	}

	log.Println("trade cron ended")
}
