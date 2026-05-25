package kline

import (
	"context"
	"time"

	"github.com/gorilla/websocket"

	policySpotOKX "github.com/jwm1rr0rb10/kline_service/app/internal/policy/spot_kline_okx"
)

type policy interface {
	Create(ctx context.Context, req policySpotOKX.CreateKlineRequest) error
}

type OkxSpotWebSocket struct {
	conn           *websocket.Conn
	policy         policy
	symbols        []string
	intervals      []string
	reconnectTimes uint8
	reconnectDelay time.Duration

	errSignal chan error
	stop      chan struct{}
	pingStop  chan struct{}
}

func NewOkxSpotWebSocket(
	policy policy,
	symbols, intervals []string,
	reconnectTimes uint8,
	reconnectDelay time.Duration,
) *OkxSpotWebSocket {
	return &OkxSpotWebSocket{
		policy:         policy,
		symbols:        symbols,
		intervals:      intervals,
		reconnectTimes: reconnectTimes,
		reconnectDelay: reconnectDelay,
		errSignal:      make(chan error, 1),
		stop:           make(chan struct{}),
		pingStop:       make(chan struct{}),
	}
}
