package service

import (
	"context"

	"github.com/jwm1rr0rb10/go-errors"
	"github.com/jwm1rr0rb10/go-logging"
	"github.com/jwm1rr0rb10/libraries/backend/golang/sfqb"

	domainKlineOKX "github.com/jwm1rr0rb10/kline_service/app/internal/domain/kline_okx"
	domainKlineOKXModel "github.com/jwm1rr0rb10/kline_service/app/internal/domain/kline_okx/model"
)

func (s *Service) Create(ctx context.Context, req *domainKlineOKXModel.Kline) error {
	logging.L(ctx).Debug("Service Create Kline Spot")

	err := s.storage.Create(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, domainKlineOKX.ErrViolatesConstraintSpotIDPK):
			return domainKlineOKX.ErrKlineAlreadyExist
		case errors.Is(err, domainKlineOKX.ErrViolatesConstraintSpotSymbolOpenCloseUnq):
			return domainKlineOKX.ErrKlineAlreadyExist
		}
		return errors.Wrap(err, "s.storage.Create")
	}

	return nil
}

func (s *Service) All(ctx context.Context, filters sfqb.SFQB) (*[]domainKlineOKXModel.Kline, error) {
	logging.L(ctx).Debug("Service All Kline Spot")

	result, err := s.storage.All(ctx, filters)
	if err != nil {
		return nil, errors.Wrap(err, "s.storage.All")
	}
	return result, nil
}

func (s *Service) Count(ctx context.Context, filters sfqb.SFQB) (uint64, error) {
	logging.L(ctx).Debug("Service Count Kline Spot")

	count, err := s.storage.Count(ctx, filters)
	if err != nil {
		return 0, errors.Wrap(err, "s.storage.Count")
	}

	return count, nil
}
