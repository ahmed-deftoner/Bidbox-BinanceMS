package db

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //postgres database driver
)

var DB *gorm.DB

func Initialize(Dbdriver, DbUser, DbPassword, DbPort, DbHost, DbName string) {

	var err error
	DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", DbHost, DbPort, DbUser, DbName, DbPassword)
	DB, err = gorm.Open(Dbdriver, DBURL)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Connected to the database")
	}

}

type CoinPair struct {
	Id       int    `json:"id"`
	Coin     string `json:"coin"`
	Active   bool
	Leverage int
}

func (cp *CoinPair) GetAllCoins() (*[]CoinPair, error) {
	Coins := []CoinPair{}
	err := DB.Model(&CoinPair{}).Where("active = ?", true).Find(&Coins).Error
	if err != nil {
		return &[]CoinPair{}, err
	}
	return &Coins, nil
}

type Key struct {
	Service     string `gorm:"size:255;not null" json:"service"`
	UserEmail   string `json:"user_email"`
	ApiKey      string `gorm:"not null;unique" json:"api_key"`
	SecretKey   string `gorm:"not null;unique" json:"secret_key"`
	Passphrase  string `gorm:"" json:"passphrase"`
	OpenShort   int
	OpenLong    int
	TradeAmount int
	Strategy    string
	Start       bool
	Mode        string
}

func (u *Key) FindAllKeys() (*[]Key, error) {
	Keys := []Key{}
	err := DB.Model(&Key{}).Where("service = ?", "binance").Find(&Keys).Error
	if err != nil {
		return &[]Key{}, err
	}
	return &Keys, nil
}

func (u *Key) FindKeyByEmail(email string) (*Key, error) {
	Keys := Key{}
	err := DB.Model(&Key{}).Where("service = ? AND email = ?", "binance", email).Take(&Keys).Error
	if err != nil {
		return &Key{}, err
	}
	return &Keys, nil
}

func (u *Key) ChangePositions(long int, short int) (*Key, error) {
	DB = DB.Model(&Key{}).Where("user_email = ? AND service = ?", u.UserEmail, u.Service).Take(&Key{}).UpdateColumns(
		map[string]interface{}{
			"open_short": short,
			"open_long":  long,
		},
	)

	if DB.Error != nil {
		return &Key{}, DB.Error
	}
	// This is the display the updated user
	err := DB.Model(&Key{}).Where("user_email = ? AND service = ?", u.UserEmail, u.Service).Take(&u).Error
	if err != nil {
		return &Key{}, err
	}
	return u, nil
}

type Conditions struct {
	Capital   int
	Positions int
	Leverage  int
	StopLoss  int
}

func (u *Conditions) FindAllConditions() (*[]Conditions, error) {
	conditions := []Conditions{}
	err := DB.Model(Conditions{}).Limit(100).Take(conditions).Error
	if err != nil {
		return &[]Conditions{}, err
	}
	return &conditions, nil
}

func (u *Conditions) FindCondition(capital int) (*Conditions, error) {
	cond := Conditions{}
	err := DB.Model(Conditions{}).Where("capital = ?", capital).Take(&cond).Error
	if err != nil {
		return &Conditions{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &Conditions{}, errors.New("Key not found")
	}
	return &cond, nil
}

type Positions struct {
	Id              int       `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	CreatedAt       time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	Symbol          string    `json:"symbol"`
	Leverage        string    `json:"leverage"`
	OpenPrice       string    `json:"open_price"`
	LiqPrice        string    `json:"liq_price"`
	TakeProfit      string    `json:"take_profit"`
	MarkPrice       string    `json:"mark_price"`
	StopLoss        string    `json:"stop_loss"`
	UnrealizedPl    string    `json:"unrealized_pl"`
	Side            string    `json:"side"`
	Size            string    `json:"size"`
	Margin          string    `json:"margin"`
	UserEmail       string    `gorm:"not null" json:"user_email"`
	Status          string    `gorm:"default:'opened'" json:"status"`
	Exchange        string    `json:"exchange"`
	LastUpdatePrice string    `json:"last_update_price"`
	OrderId         string
	Layer           int
	FirstBuyAmount  string
}

func (position *Positions) CreateNewPosition() (*Positions, error) {
	err := DB.Create(&position).Error
	if err != nil {
		return &Positions{}, err
	}
	return position, nil
}

func (position *Positions) FindAllPositions() (*[]Positions, error) {
	pos := []Positions{}
	err := DB.Model(&Positions{}).Where("exchange = ? AND status = ?", "binance", "opened").Find(&pos).Error
	if err != nil {
		return &[]Positions{}, err
	}
	return &pos, nil
}

func (position *Positions) FindAllUserPositions(email string) (*[]Positions, error) {
	pos := []Positions{}
	err := DB.Model(&Positions{}).Where("exchange = ? AND status = ? AND user_email = ?", "binance", "opened", email).Find(&pos).Error
	if err != nil {
		log.Println("ddd")
		log.Println(err)
		return &[]Positions{}, err
	}
	return &pos, nil
}

func (position *Positions) UpdateLayer(side string, symbol string, size string, price float64) error {
	fsize, err := strconv.ParseFloat(size, 64)
	if err != nil {
		log.Fatal(err)
	}

	dsize, err := strconv.ParseFloat(position.Size, 64)
	if err != nil {
		log.Fatal(err)
	}
	x := fsize + dsize
	pos := []Positions{}
	err = DB.Model(&Positions{}).Where("exchange = ? AND side = ?", "bybit", side).Find(&pos).Error
	if err != nil {
		return err
	}
	DB = DB.Model(&Positions{}).Where("user_email = ? AND exchange = ? AND symbol = ? AND side = ?", position.UserEmail, "bybit", symbol, side).Take(&Positions{}).UpdateColumns(
		map[string]interface{}{
			"layer": pos[0].Layer + 1,
			"size":  fmt.Sprintf("%.3f", x),
		},
	)

	if DB.Error != nil {
		return DB.Error
	}

	return nil
}

func (position *Positions) UpdatePrice(symbol string, price string) error {
	DB = DB.Model(&Positions{}).Where("exchange = ? AND symbol = ?", "bybit", symbol).UpdateColumns(
		map[string]interface{}{
			"open_price": price,
		},
	)

	if DB.Error != nil {
		return DB.Error
	}

	return nil
}

func (position *Positions) UpdateStatus(symbol string) error {
	DB = DB.Model(&Positions{}).Where("exchange = ? AND symbol = ?", "bybit", symbol).UpdateColumns(
		map[string]interface{}{
			"status": "closed",
		},
	)

	if DB.Error != nil {
		return DB.Error
	}

	return nil
}

type BybitOrderRequest struct {
	Category       string `json:"category"`
	Symbol         string `json:"symbol"`
	Side           string `json:"side"`
	OrderType      string `json:"orderType"`
	Qty            string `json:"qty"`
	TimeInForce    string `json:"timeInForce"`
	ReduceOnly     bool   `json:"reduce_only"`
	CloseOnTrigger bool   `json:"closeOnTrigger"`
	TakeProfit     string `json:"takeProfit"`
	StopLoss       string `json:"stopLoss"`
	SlTriggerBy    string `json:"slTriggerBy"`
	TpTriggerBy    string `json:"tpTriggerBy"`
	TpslMode       string `json:"tpslMode"`
	SlOrderType    string `json:"slOrderType"`
	SlLimitPrice   string `json:"slLimitPrice"`
}
type BybitResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		OrderID     string `json:"orderId"`
		OrderLinkId string `json:"orderLinkId"`
	} `json:"result"`
	RetExtInfo map[string]interface{} `json:"retExtInfo"`
	Time       int64                  `json:"time"`
}

// type BybitResponse struct {
// 	RetCode    int                    `json:"retCode"`
// 	RetMsg     string                 `json:"retMsg"`
// 	Result     struct {
// 		OrderID     string `json:"order_id"`
// 		OrderLinkId string `json:"orderLinkId"`
// 	}
// 	RetExtInfo map[string]interface{} `json:"retExtInfo"`
// 	Time       int64                  `json:"time"`
// }

type OrderRequest struct {
	Symbol     string `json:"symbol"`
	MarginCoin string `json:"marginCoin"`
	Size       string `json:"size"`
	Side       string `json:"side"`
	OrderType  string `json:"orderType"`
	StopLoss   string `json:"presetStopLossPrice"`
	TakeProfit string `json:"presetTakeProfitPrice"`
}

type TriggerType string
type OrderSide string

const (
	FillPrice   TriggerType = "fill_price"
	MarketPrice TriggerType = "market_price"
)

const (
	OpenLong   OrderSide = "open_long"
	OpenShort  OrderSide = "open_short"
	CloseLong  OrderSide = "close_long"
	CloseShort OrderSide = "close_short"
	BuySingle  OrderSide = "buy_single"
	SellSingle OrderSide = "sell_single"
)

type TrailingStopOrderRequest struct {
	Symbol       string      `json:"symbol"`
	MarginCoin   string      `json:"marginCoin"`
	TriggerPrice string      `json:"triggerPrice"`
	TriggerType  TriggerType `json:"triggerType"`
	Size         string      `json:"size"`
	Side         string      `json:"side"`
	RangeRate    string      `json:"rangeRate"`
}

type BybitTrailingStopOrderRequest struct {
	Category     string `json:"category"`
	Symbol       string `json:"symbol"`
	TakeProfit   string `json:"takeProfit"`
	StopLoss     string `json:"stopLoss"`
	TrailingStop string `json:"trailingStop"`
	ActivePrice  string `json:"activePrice"`
	TpslMode     string `json:"tpslMode"`
	TpSize       string `json:"tpSize"`
	SlSize       string `json:"slSize"`
	TpOrderType  string `json:"tpOrderType"`
	SlOrderType  string `json:"slOrderType"`
	TpTriggerBy  string `json:"tpTriggerBy"`
	SlTriggerBy  string `json:"slTriggerBy"`
	TpLimitPrice string `json:"tpLimitPrice"`
	SlLimitPrice string `json:"slLimitPrice"`
	PositionIdx  int    `json:"positionIdx"`
}

type OrderResponse struct {
	Code        string `json:"code"`
	Msg         string `json:"msg"`
	RequestTime int64  `json:"requestTime"`
	Data        struct {
		ClientOid string `json:"clientOid"`
		OrderID   string `json:"orderId"`
	} `json:"data"`
}

type Order struct {
	Email       string
	Symbol      string
	MarginCoin  string
	Size        string
	Side        string
	OrderType   string
	Service     string
	QuoteAmount float64
	Profit      float64
}

func (o *Order) Initialize(order OrderRequest, email string, client_id string, order_id string) {
	o.MarginCoin = order.MarginCoin
	o.Side = order.Side
	o.Symbol = order.Symbol
	o.Size = order.Size
	o.OrderType = order.OrderType
	o.Email = email
}

func (o *OrderRequest) Validate() error {
	if o.MarginCoin == "" {
		return errors.New("margin coin is required")
	}
	if o.OrderType == "" {
		return errors.New("ordertype is required")
	}
	if o.Side == "" {
		return errors.New("side is required")
	}
	if o.Size == "" {
		return errors.New("size is required")
	}
	if o.Symbol == "" {
		return errors.New("symbol is required")
	}
	return nil
}

func (o *Order) SaveOrder() (*Order, error) {
	err := DB.Create(&o).Error
	if err != nil {
		return &Order{}, err
	}
	return o, nil
}
