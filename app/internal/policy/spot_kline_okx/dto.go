package spot_kline_okx

import (
	domainKlineOKXModel "github.com/jwm1rr0rb10/kline_service/app/internal/domain/kline_okx/model"
	"github.com/jwm1rr0rb10/libraries/backend/golang/sfqb"
)

type CreateKlineRequest struct {
	StartTime        int64  `json:"start_time"         db:"start_time"`
	CloseTime        int64  `json:"close_time"         db:"close_time"`
	OpenPrice        string `json:"open_price"         db:"open_price"`
	ClosePrice       string `json:"close_price"        db:"close_price"`
	HighPrice        string `json:"high_price"         db:"high_price"`
	LowPrice         string `json:"low_price"          db:"low_price"`
	BaseAssetVolume  string `json:"base_asset_volume"  db:"base_asset_volume"`
	QuoteAssetVolume string `json:"quote_asset_volume" db:"quote_asset_volume"`
	Symbol           string `json:"symbol"             db:"symbol"`
	Interval         string `json:"interval"           db:"interval"`
}

type SearchKlineRequest struct {
	Filters   sfqb.SFQB `json:"filters"`
	WithCount bool      `json:"with_count"`
}

func NewSearchKlineRequest(
	filters sfqb.SFQB,
	withCount bool,
) *SearchKlineRequest {
	return &SearchKlineRequest{
		Filters:   filters,
		WithCount: withCount,
	}
}

type SearchKlineResponse struct {
	Items *[]domainKlineOKXModel.Kline `json:"items"`
	Count *uint64                      `json:"count"`
}

func NewSearchKlineResponse(
	items *[]domainKlineOKXModel.Kline,
	count *uint64,
) *SearchKlineResponse {
	return &SearchKlineResponse{
		Items: items,
		Count: count,
	}
}
