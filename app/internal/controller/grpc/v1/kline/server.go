package kline

import (
	"context"

	gRPCKlineService "github.com/jwm1rr0rb10/kline_contract/gen/go/kline_service"

	policySpotOKX "github.com/jwm1rr0rb10/kline_service/app/internal/policy/spot_kline_okx"
)

type policy interface {
	SearchKline(context.Context, policySpotOKX.SearchKlineRequest) (*policySpotOKX.SearchKlineResponse, error)
}

type Controller struct {
	gRPCKlineService.UnimplementedKlineServiceServer
}
