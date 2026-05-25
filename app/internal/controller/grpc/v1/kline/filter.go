package kline

import (
	"github.com/jwm1rr0rb10/go-errors"
	gRPCKlineService "github.com/jwm1rr0rb10/kline_contract/gen/go/kline_service/v1"
	"github.com/jwm1rr0rb10/libraries/backend/golang/apperror"
	"github.com/jwm1rr0rb10/libraries/backend/golang/queryify"
	"github.com/jwm1rr0rb10/libraries/backend/golang/sfqb"

	"github.com/jwm1rr0rb10/kline_service/app/internal/dal/postgres"
	"github.com/jwm1rr0rb10/kline_service/app/internal/domain"
)

const (
	domainName = "kline"
)

const (
	maxLimit     = 1000
	defaultLimit = 100
)

const (
	validationErrCode = iota + 100
	minSearchErrCode
)

const (
	fieldNameID          = "spot.id"
	fieldNameSymbol      = "spot.symbol"
	fieldNameInterval    = "spot.interval"
	fieldNameOpenTime    = "spot.open_time"
	fieldNameCloseTime   = "spot.close_time"
	fieldNameTypeTrend   = "spot.type_trend"
	fieldNameOpenPrice   = "spot.open_price"
	fieldNameHighPrice   = "spot.high_price"
	fieldNameLowPrice    = "spot.low_price"
	fieldNameClosePrice  = "spot.close_price"
	fieldNameBaseVolume  = "spot.base_volume"
	fieldNameQuoteVolume = "spot.quote_volume"
)

func buildValidationSearchKlineFilters(req *gRPCKlineService.SearchKlineRequest) (sfqb.SFQB, error) {
	searchFields := []string{
		fieldNameID,
		fieldNameSymbol,
		fieldNameInterval,
		fieldNameOpenTime,
		fieldNameCloseTime,
		fieldNameTypeTrend,
		fieldNameOpenPrice,
		fieldNameHighPrice,
		fieldNameLowPrice,
		fieldNameClosePrice,
		fieldNameBaseVolume,
		fieldNameQuoteVolume,
	}

	errFields := apperror.ErrorFields{}

	filters, err := queryify.NewFilters(
		queryify.WithSearchFields(searchFields),
		// need rewrite my custom Lib with new contract(common from kline_contract). Now i use my common contract
		// what I use in my projects.
		//queryify.WithPaginator(req, defaultLimit, maxLimit),
		//queryify.WithSorter(req),
		queryify.WithSearcher(req),
		queryify.WithMinSearchSymbols(postgres.SearchMinSymbols),
	)
	if err != nil {
		if errors.Is(err, queryify.ErrSearchMinSymbols) {
			return nil, apperror.NewValidationError(
				domain.SystemCode,
				apperror.WithDomain(domainName),
				apperror.WithMessage("search minimum symbols error"),
				apperror.WithCode(minSearchErrCode),
			)
		}

		return nil, errors.Wrap(err, "queryify.NewFilters")
	}

	var operator sfqb.Method

	var parseErr error

	// 12 if !DRY because here need using GetOp for a lot of type of data.
	// need a little bit of change my lib queryify.
	if kLineIDVal := req.GetKlineIdVal(); kLineIDVal != nil {
		operator, parseErr = queryify.MapOperator(kLineIDVal.GetOp())
		if parseErr != nil {
			errFields[fieldNameID] = parseErr.Error()
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNameID,
				Method: operator,
				Value:  kLineIDVal.GetVal(),
			})
		}
	}

	if symbolVal := req.GetSymbolVal(); symbolVal != nil {
		operator, parseErr = queryify.MapOperator(symbolVal.GetOp())
		if parseErr != nil {
			errFields[fieldNameSymbol] = parseErr.Error()
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNameSymbol,
				Method: operator,
				Value:  symbolVal.GetVal(),
			})
		}
	}

	if intervalVal := req.GetIntervalVal(); intervalVal != nil {
		operator, parseErr = queryify.MapOperator(intervalVal.GetOp())
		if parseErr != nil {
			errFields[fieldNameInterval] = parseErr.Error()
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNameInterval,
				Method: operator,
				Value:  intervalVal.GetVal(),
			})
		}
	}

	if openPriceVal := req.GetOpenPriceVal(); openPriceVal != nil {
		operator, parseErr = queryify.MapOperator(openPriceVal.GetOp())
		if parseErr != nil {
			errFields[fieldNameOpenPrice] = parseErr.Error()
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNameOpenPrice,
				Method: operator,
				Value:  openPriceVal.GetVal(),
			})
		}
	}

	if highPriceVal := req.GetHighPriceVal(); highPriceVal != nil {
		operator, parseErr = queryify.MapOperator(highPriceVal.GetOp())
		if parseErr != nil {
			errFields[fieldNameHighPrice] = parseErr.Error()
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNameHighPrice,
				Method: operator,
				Value:  highPriceVal.GetVal(),
			})
		}
	}

	if lowPriceVal := req.GetLowPriceVal(); lowPriceVal != nil {
		operator, parseErr = queryify.MapOperator(lowPriceVal.GetOp())
		if parseErr != nil {
			errFields[fieldNameLowPrice] = parseErr.Error()
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNameLowPrice,
				Method: operator,
				Value:  lowPriceVal.GetVal(),
			})
		}
	}

	if closePriceVal := req.GetClosePriceVal(); closePriceVal != nil {
		operator, parseErr = queryify.MapOperator(closePriceVal.GetOp())
		if parseErr != nil {
			errFields[fieldNameClosePrice] = parseErr.Error()
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNameClosePrice,
				Method: operator,
				Value:  closePriceVal.GetVal(),
			})
		}
	}

	if baseVolumeVal := req.GetBaseAssetVolumeVal(); baseVolumeVal != nil {
		operator, parseErr = queryify.MapOperator(baseVolumeVal.GetOp())
		if parseErr != nil {
			errFields[fieldNameBaseVolume] = parseErr.Error()
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNameBaseVolume,
				Method: operator,
				Value:  baseVolumeVal.GetVal(),
			})
		}
	}

	if quoteVolumeVal := req.GetQuoteAssetVolumeVal(); quoteVolumeVal != nil {
		operator, parseErr = queryify.MapOperator(quoteVolumeVal.GetOp())
		if parseErr != nil {
			errFields[fieldNameQuoteVolume] = parseErr.Error()
		} else {
			filters.AddFilter(sfqb.FilterField{
				Name:   fieldNameQuoteVolume,
				Method: operator,
				Value:  quoteVolumeVal.GetVal(),
			})
		}
	}

	if len(errFields) > 0 {
		return nil, apperror.NewValidationError(
			domain.SystemCode,
			apperror.WithDomain(domainName),
			apperror.WithMessage("validation error"),
			apperror.WithFields(errFields),
			apperror.WithCode(validationErrCode),
		)
	}

	return filters, nil
}
