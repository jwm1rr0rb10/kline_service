package kline

import "context"

type policy interface {
	SearchKline(ctx context.Context, policy)
}
