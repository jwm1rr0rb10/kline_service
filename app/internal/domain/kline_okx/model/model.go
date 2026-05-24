package model

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/jwm1rr0rb10/go-errors"
)

var zeroDecimal = decimal.NewFromInt(0)

type Kline struct {
	StartTime        time.Time       `json:"start_time"         db:"start_time"`
	CloseTime        time.Time       `json:"close_time"         db:"close_time"`
	OpenPrice        decimal.Decimal `json:"open_price"         db:"open_price"`
	ClosePrice       decimal.Decimal `json:"close_price"        db:"close_price"`
	HighPrice        decimal.Decimal `json:"high_price"         db:"high_price"`
	LowPrice         decimal.Decimal `json:"low_price"          db:"low_price"`
	BaseAssetVolume  decimal.Decimal `json:"base_asset_volume"  db:"base_asset_volume"`
	QuoteAssetVolume decimal.Decimal `json:"quote_asset_volume" db:"quote_asset_volume"`
	ID               string          `json:"id"                 db:"id"`
	Symbol           string          `json:"symbol"             db:"symbol"`
	Interval         string          `json:"interval"           db:"interval"`
	TypeTrend        bool            `json:"type_trend"         db:"type_trend"`
}

type KlineBuilder struct {
	kline Kline
}

func NewKlineBuilder() *KlineBuilder {
	return &KlineBuilder{}
}

func (b *KlineBuilder) Reset() *KlineBuilder {
	b.kline = Kline{}
	return b
}

func (b *KlineBuilder) WithTimes(start, close time.Time) *KlineBuilder {
	b.kline.StartTime = start
	b.kline.CloseTime = close
	return b
}

func (b *KlineBuilder) WithPrices(open, close, high, low decimal.Decimal) *KlineBuilder {
	b.kline.OpenPrice = open
	b.kline.ClosePrice = close
	b.kline.HighPrice = high
	b.kline.LowPrice = low
	return b
}

func (b *KlineBuilder) WithVolumes(base, quote decimal.Decimal) *KlineBuilder {
	b.kline.BaseAssetVolume = base
	b.kline.QuoteAssetVolume = quote
	return b
}

func (b *KlineBuilder) WithSymbolInfo(id, symbol, interval string) *KlineBuilder {
	b.kline.ID = id
	b.kline.Symbol = symbol
	b.kline.Interval = interval
	return b
}

func (b *KlineBuilder) WithTrend(typeTrend bool) *KlineBuilder {
	b.kline.TypeTrend = typeTrend
	return b
}

func (b *KlineBuilder) Build() (Kline, error) {
	if err := b.validate(); err != nil {
		return Kline{}, err
	}
	return b.kline, nil
}

func (b *KlineBuilder) validate() error {
	var errs error

	if err := b.validateTime(); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := b.validateRequiredFields(); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := b.validateNonNegative(); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := b.validatePriceLogic(); err != nil {
		errs = errors.Append(errs, err)
	}

	return errs
}

func (b *KlineBuilder) validateTime() error {
	if b.kline.CloseTime.Before(b.kline.StartTime) {
		return errors.New("closeTime must be after startTime")
	}
	return nil
}

func (b *KlineBuilder) validateRequiredFields() error {
	if b.kline.Symbol == "" || b.kline.Interval == "" {
		return errors.New("symbol and interval are required")
	}
	return nil
}

func (b *KlineBuilder) validateNonNegative() error {
	if b.kline.OpenPrice.LessThan(zeroDecimal) || b.kline.ClosePrice.LessThan(zeroDecimal) ||
		b.kline.HighPrice.LessThan(zeroDecimal) || b.kline.LowPrice.LessThan(zeroDecimal) {
		return errors.New("prices cannot be negative")
	}
	if b.kline.BaseAssetVolume.LessThan(zeroDecimal) || b.kline.QuoteAssetVolume.LessThan(zeroDecimal) {
		return errors.New("volumes cannot be negative")
	}
	return nil
}

func (b *KlineBuilder) validatePriceLogic() error {
	if b.kline.HighPrice.LessThan(b.kline.LowPrice) {
		return errors.New("high price cannot be less than low price")
	}
	if b.kline.HighPrice.LessThan(b.kline.OpenPrice) || b.kline.HighPrice.LessThan(b.kline.ClosePrice) {
		return errors.New("high price must be the highest price in the spot_kline_okx")
	}
	if b.kline.LowPrice.GreaterThan(b.kline.OpenPrice) || b.kline.LowPrice.GreaterThan(b.kline.ClosePrice) {
		return errors.New("low price must be the lowest price in the spot_kline_okx")
	}
	return nil
}
