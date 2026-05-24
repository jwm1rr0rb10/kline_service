package service

import (
	"context"

	"github.com/jwm1rr0rb10/libraries/backend/golang/sfqb"

	domainKlineOKXModel "github.com/jwm1rr0rb10/kline_service/app/internal/domain/kline_okx/model"
)

type storage interface {
	Create(context.Context, *domainKlineOKXModel.Kline) error
	All(context.Context, sfqb.SFQB) (*[]domainKlineOKXModel.Kline, error)
	Count(context.Context, sfqb.SFQB) (uint64, error)
}

type Service struct {
	storage storage
}

func New(storage storage) *Service {
	return &Service{
		storage: storage,
	}
}
