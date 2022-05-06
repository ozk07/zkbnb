package price

import (
	"encoding/json"
	"fmt"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/l2asset"
	"io/ioutil"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var (
	cachePriceSymbolPrefix = "cache::price:symbol:"
)

type (
	PriceModel interface {
		UpdateCurrencyPrice() error
		UpdateCurrencyPriceBySymbol(symbol string) error
		GetCurrencyPrice(currency string) (price float64, err error)
	}

	defaultPriceModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	Price struct {
		gorm.Model
	}
)

func NewPriceModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) PriceModel {
	return &defaultPriceModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      `price`,
		DB:         db,
	}
}

func (*Price) TableName() string {
	return `price`
}

func GetQuotesLatest(l2Symbol string, client *http.Client) (quotesLatest []*QuoteLatest, err error) {
	currency := l2Symbol
	url := fmt.Sprintf("%s%s", CoinMarketCap, currency)

	// Get Request
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logx.Error("[price] New Request Error %s", err.Error())
		return nil, err
	}

	// Add Header
	reqest.Header.Add("X-CMC_PRO_API_KEY", "cfce503f-dd3d-4847-9570-bbab5257dac8")
	reqest.Header.Add("Accept", "application/json")

	resp, err := client.Do(reqest)
	if err != nil {
		errInfo := fmt.Sprintf("[price] Network Error %s", err.Error())
		logx.Error(errInfo)
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	currencyPrice := new(CurrencyPrice)
	err = json.Unmarshal(body, &currencyPrice)
	if err != nil {
		errInfo := fmt.Errorf("[price] JSON Error: [%s]. Response body: [%s]", err.Error(), string(body))
		logx.Error(errInfo)
		return nil, err
	}

	ifcs, ok := currencyPrice.Data.(interface{})
	if !ok {
		errInfo := fmt.Sprintf("[price] %s", ErrTypeAssertion)
		logx.Error(errInfo)
		return nil, ErrTypeAssertion
	}

	for _, coinObj := range ifcs.(map[string]interface{}) {
		quoteLatest := new(QuoteLatest)
		b, err := json.Marshal(coinObj)
		if err != nil {
			logx.Error("[price] Marshal Error")
			return nil, err
		}

		err = json.Unmarshal(b, quoteLatest)
		if err != nil {
			logx.Error("[price] Unmarshal Error")
			return nil, err
		}

		quotesLatest = append(quotesLatest, quoteLatest)
	}

	return quotesLatest, nil
}

/*
	Func: UpdateCurrencyPrice
	Params:
	Return: err
	Description: update currency price cache
*/
func (m *defaultPriceModel) UpdateCurrencyPrice() error {
	myClient := &http.Client{}

	var (
		l2AssetTable = "l2_asset_info"
		l2AssetInfos []*l2asset.L2AssetInfo
	)
	dbTx := m.DB.Table(l2AssetTable).Find(&l2AssetInfos)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[price.GetL2AssetsList] %s", dbTx.Error)
		logx.Error(err)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[price.GetL2AssetsList] %s", ErrNotFound)
		logx.Error(err)
		return ErrNotFound
	}

	var l2Symbol string
	for i := 0; i < len(l2AssetInfos); i++ {
		// REY IS NOT YET
		if l2AssetInfos[i].AssetSymbol == "REY" {
			continue
		}
		if len(l2Symbol) == 0 {
			l2Symbol += l2AssetInfos[i].AssetSymbol
		} else {
			l2Symbol += "," + l2AssetInfos[i].AssetSymbol
		}
	}

	quotesLatest, err := GetQuotesLatest(l2Symbol, myClient)
	if err != nil {
		errInfo := fmt.Sprintf("[PriceModel.UpdatePrice.GetQuotesLatest] %s", err)
		logx.Error(errInfo)
		return err
	}

	for _, quoteLatest := range quotesLatest {
		key := fmt.Sprintf("%s%v", cachePriceSymbolPrefix, quoteLatest.Symbol)

		if quoteLatest.Quote["USD"] != nil {
			err = m.SetCache(key, quoteLatest.Quote["USD"].Price)
			if err != nil {
				errInfo := fmt.Sprintf("[PriceModel.UpdatePrice.Setcache] %s", err)
				logx.Error(errInfo)
				return err
			}

			logx.Info(fmt.Sprintf("Currency:%s, Price:%+v", quoteLatest.Symbol, quoteLatest.Quote["USD"].Price))
		} else {
			errInfo := fmt.Sprintf("[PriceModel.UpdatePrice] get %s usd price from coinmarketcap failed", quoteLatest.Symbol)
			logx.Error(errInfo)
		}
	}

	// set REYUSDT to 0.8
	key := fmt.Sprintf("%s%v", cachePriceSymbolPrefix, "REY")
	err = m.SetCache(key, 0.8)
	if err != nil {
		errInfo := fmt.Sprintf("[PriceModel.UpdatePrice.Setcache] %s", err)
		logx.Error(errInfo)
		return err
	}

	return nil
}

/*
	Func: UpdateCurrencyPriceBySymbol
	Params:
	Return: err
	Description: update currency price cache by symbol
*/
func (m *defaultPriceModel) UpdateCurrencyPriceBySymbol(symbol string) error {
	// // proxy server setup
	// dialSocksProxy, err := proxy.SOCKS5("tcp", "172.30.144.1:7890", nil, proxy.Direct)
	// if err != nil {
	// 	fmt.Println("Error connecting to proxy:", err)
	// }
	// tr := &http.Transport{Dial: dialSocksProxy.Dial}

	// // Create client
	// myClient := &http.Client{
	// 	Transport: tr,
	// }

	myClient := &http.Client{}

	quotesLatest, err := GetQuotesLatest(symbol, myClient)
	if err != nil {
		errInfo := fmt.Sprintf("[PriceModel.UpdatePrice.GetQuotesLatest] %s", err)
		logx.Error(errInfo)
		return err
	}

	for _, quoteLatest := range quotesLatest {
		key := fmt.Sprintf("%s%v", cachePriceSymbolPrefix, quoteLatest.Symbol)
		err = m.SetCache(key, quoteLatest.Quote["USD"].Price)
		if err != nil {
			errInfo := fmt.Sprintf("[PriceModel.UpdatePrice.Setcache] %s", err)
			logx.Error(errInfo)
			return err
		}

		logx.Info(fmt.Sprintf("%+v", quoteLatest.Quote["USD"].Price))
	}

	return nil
}

/*
	Func: GetCurrencyPrice
	Params: currency string
	Return: price float64, err error
	Description: get currency price cache by currency symbol
*/
func (m *defaultPriceModel) GetCurrencyPrice(currency string) (price float64, err error) {
	key := fmt.Sprintf("%s%v", cachePriceSymbolPrefix, currency)
	err = m.QueryRow(&price, key, func(conn sqlx.SqlConn, v interface{}) error {
		return ErrNotFound
	})
	if err != nil {
		errInfo := fmt.Sprintf("[PriceModel.GetCurrencyPrice.Getcache] %s %s", key, err)
		logx.Error(errInfo)
		return 0, err
	}
	return price, nil
}
