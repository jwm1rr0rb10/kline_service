package kline

import (
	"context"

	gRPCKlineService "github.com/jwm1rr0rb10/kline_contract/gen/go/kline_service/v1"

	policySpotOKX "github.com/jwm1rr0rb10/kline_service/app/internal/policy/spot_kline_okx"
)

// better using gRPC Striming( Sorry for it) I had a little time for this test task, but
// if we're working with big data and need Search better using strimming than Unary.
type policy interface {
	SearchKline(context.Context, *policySpotOKX.SearchKlineRequest) (*policySpotOKX.SearchKlineResponse, error)
}

type Controller struct {
	gRPCKlineService.UnimplementedKlineServiceServer
	policy policy
}

func NewController(policy policy) *Controller {
	return &Controller{
		policy: policy,
	}
}
