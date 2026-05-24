package spot_kline_okx

import (
	"context"
	"time"

	"github.com/jwm1rr0rb10/go-errors"
	"github.com/jwm1rr0rb10/go-logging"
	"github.com/jwm1rr0rb10/go-tracing"
	domainKlineOKXModel "github.com/jwm1rr0rb10/kline_service/app/internal/domain/kline_okx/model"
	"github.com/shopspring/decimal"
)

func (p *Policy) Create(ctx context.Context, req CreateKlineRequest) error {
	logging.L(ctx).Debug("Policy Create Kline Spot")

	typeTrend := req.ClosePrice >= req.OpenPrice

	openTime := time.UnixMilli(req.StartTime)
	closeTime := time.UnixMilli(req.CloseTime)

	openPrice, openPriceErr := decimal.NewFromString(req.OpenPrice)
	if openPriceErr != nil {
		return errors.Wrap(openPriceErr, "invalid openPrice")
	}
	highPrice, highPriceErr := decimal.NewFromString(req.HighPrice)
	if highPriceErr != nil {
		return errors.Wrap(highPriceErr, "invalid highPrice")
	}
	lowPrice, lowPriceErr := decimal.NewFromString(req.LowPrice)
	if lowPriceErr != nil {
		return errors.Wrap(lowPriceErr, "invalid lowPrice")
	}
	closePrice, closePriceErr := decimal.NewFromString(req.ClosePrice)
	if closePriceErr != nil {
		return errors.Wrap(closePriceErr, "invalid closePrice")
	}
	baseAssetVolume, baseAssetVolumeErr := decimal.NewFromString(req.BaseAssetVolume)
	if baseAssetVolumeErr != nil {
		return errors.Wrap(baseAssetVolumeErr, "invalid baseAssetVolume")
	}
	quoteAssetVolume, quoteAssetVolumeErr := decimal.NewFromString(req.QuoteAssetVolume)
	if quoteAssetVolumeErr != nil {
		return errors.Wrap(quoteAssetVolumeErr, "invalid quoteAssetVolume")
	}

	kline, klineErr := domainKlineOKXModel.NewKlineBuilder().
		WithTimes(openTime, closeTime).
		WithPrices(openPrice, closePrice, highPrice, lowPrice).
		WithVolumes(baseAssetVolume, quoteAssetVolume).
		WithSymbolInfo(p.BasePolicy.Generate(), req.Symbol, req.Interval).
		WithTrend(typeTrend).
		Build()

	if klineErr != nil {
		return errors.Wrap(klineErr, "kline builder validation failed")
	}

	if err := p.service.Create(ctx, &kline); err != nil {
		return errors.Wrap(err, "p.service.Create")
	}

	return nil
}

func (p *Policy) SearchKline(ctx context.Context, req SearchKlineRequest) (*SearchKlineResponse, error) {
	ctx, span := tracing.Continue(ctx, "spotKlineOKXPolicy.SearchKline")
	defer span.End()

	tracing.TraceValue(ctx, "filters", req.Filters)

	logging.L(ctx).Debug("spotKlineOKXPolicy.SearchKline")

	all, err := p.service.All(ctx, req.Filters)
	if err != nil {
		return nil, errors.Wrap(err, "p.service.All")
	}

	var count *uint64

	if req.WithCount {
		kCount, kCountErr := p.service.Count(ctx, req.Filters)
		if kCountErr != nil {
			return nil, errors.Wrap(kCountErr, "p.service.Count")
		}

		count = &kCount
	}

	response := NewSearchKlineResponse(
		all,
		count,
	)

	return response, nil
}
