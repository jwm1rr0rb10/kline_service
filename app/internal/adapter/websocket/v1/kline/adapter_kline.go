package kline

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jwm1rr0rb10/libraries/backend/golang/logging"

	policySpotOKX "github.com/jwm1rr0rb10/kline_service/app/internal/policy/spot_kline_okx"
)

func (sws *OkxSpotWebSocket) Connect(ctx context.Context) error {
	var reconnectCount uint8

	for {
		c, _, err := websocket.DefaultDialer.Dial(okxWSURL, nil)
		if err != nil {
			reconnectCount++
			logging.L(ctx).Error("OKX WebSocket dial failed", logging.AnyAttr("attempt", reconnectCount), logging.ErrAttr(err))

			if reconnectCount >= sws.reconnectTimes {
				return fmt.Errorf("too many reconnection attempts: %w", err)
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(sws.reconnectDelay):
				continue
			}
		}
		sws.conn = c
		break
	}

	args := make([]map[string]string, 0, len(sws.symbols)*len(sws.intervals))
	for _, symbol := range sws.symbols {
		for _, interval := range sws.intervals {
			channelName := "candle" + normalizeOKXInterval(interval)
			inst := normalizeInstId(symbol)

			args = append(args, map[string]string{
				"channel": channelName,
				"instId":  inst,
			})
		}
	}

	subscribeMsg := map[string]interface{}{
		"op":   "subscribe",
		"args": args,
	}

	if err := sws.conn.WriteJSON(subscribeMsg); err != nil {
		return fmt.Errorf("subscribe failed: %w", err)
	}

	go sws.keepAlive(ctx)
	logging.L(ctx).Info("OKX WebSocket connected and subscribed")
	return nil
}

func (sws *OkxSpotWebSocket) Listen(ctx context.Context) {
	defer func() {
		if sws.conn != nil {
			sws.conn.Close()
		}
	}()

	for {
		select {
		case <-sws.stop:
			return
		default:
			_, message, err := sws.conn.ReadMessage()
			if err != nil {
				select {
				case sws.errSignal <- fmt.Errorf("websocket read error: %w", err):
				default:
				}
				return
			}
			sws.processKlineMessage(ctx, message)
		}
	}
}

func (sws *OkxSpotWebSocket) processKlineMessage(ctx context.Context, message []byte) {
	if len(message) == 0 || message[0] != '{' {
		return
	}

	var msg OkxMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		logging.L(ctx).Error("Failed to unmarshal OKX message", logging.ErrAttr(err))
		return
	}

	if msg.Event == "error" || msg.Event == "subscribe" || len(msg.Data) == 0 {
		return
	}

	d := msg.Data[0]
	if len(d) < 9 || d[8] != "1" {
		return
	}

	startTime, _ := strconv.ParseInt(d[0], 10, 64)
	intervalStr := strings.TrimPrefix(msg.Arg.Channel, "candle")
	closeTime := startTime + getCandleDurationMs(intervalStr)

	kline := policySpotOKX.CreateKlineRequest{
		Symbol:           msg.Arg.InstId,
		Interval:         intervalStr,
		StartTime:        startTime,
		CloseTime:        closeTime,
		OpenPrice:        d[1],
		ClosePrice:       d[4],
		HighPrice:        d[2],
		LowPrice:         d[3],
		BaseAssetVolume:  d[5],
		QuoteAssetVolume: d[6],
	}

	if err := sws.policy.Create(ctx, kline); err != nil {
		logging.L(ctx).Error("Failed to save kline from websocket", logging.ErrAttr(err))
	}
}

func (sws *OkxSpotWebSocket) Close() {
	close(sws.stop)
	close(sws.pingStop)

	if sws.conn != nil {
		_ = sws.conn.Close()
		sws.conn = nil
	}
}
