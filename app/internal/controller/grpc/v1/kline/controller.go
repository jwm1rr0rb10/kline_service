package kline

import (
	"context"

	gRPCKlineService "github.com/jwm1rr0rb10/kline_contract/gen/go/kline_service"
)

func (c *Controller) SearchKline(
	ctx context.Context,
	req gRPCKlineService.Sea)
