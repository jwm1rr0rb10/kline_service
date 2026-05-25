package kline

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	gRPCKlineService "github.com/jwm1rr0rb10/kline_contract/gen/go/kline_service/v1"
	domainKlineOKXModel "github.com/jwm1rr0rb10/kline_service/app/internal/domain/kline_okx/model"
	"github.com/jwm1rr0rb10/kline_service/app/internal/policy/spot_kline_okx"
)

func convertSearchKlineToPB(
	data *spot_kline_okx.SearchKlineResponse,
) *gRPCKlineService.SearchKlineResponse {

	if data == nil || data.Items == nil {
		return &gRPCKlineService.SearchKlineResponse{}
	}

	items := *data.Items
	response := make([]*gRPCKlineService.Kline, len(items))

	for i := range items {
		response[i] = NewKlineToPB(&items[i])
	}

	return &gRPCKlineService.SearchKlineResponse{
		Items: response,
		Count: data.Count,
	}
}

func NewKlineToPB(kline *domainKlineOKXModel.Kline) *gRPCKlineService.Kline {
	if kline == nil {
		return nil
	}

	return &gRPCKlineService.Kline{
		KlineId:          kline.ID,
		Symbol:           kline.Symbol,
		Interval:         kline.Interval,
		StartTime:        timestamppb.New(kline.StartTime),
		CloseTime:        timestamppb.New(kline.CloseTime),
		OpenPrice:        kline.OpenPrice.String(),
		ClosePrice:       kline.ClosePrice.String(),
		HighPrice:        kline.HighPrice.String(),
		LowPrice:         kline.LowPrice.String(),
		BaseAssetVolume:  kline.BaseAssetVolume.String(),
		QuoteAssetVolume: kline.QuoteAssetVolume.String(),
	}
}
