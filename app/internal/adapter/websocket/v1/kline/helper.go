package kline

import (
	"context"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const okxWSURL = "wss://ws.okx.com:8443/ws/v5/business"

func normalizeInstId(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	if strings.Contains(s, "-") {
		return s
	}
	quotes := []string{"USDT", "USDC", "USD", "BTC", "ETH", "BNB"}
	for _, q := range quotes {
		if strings.HasSuffix(s, q) {
			return s[:len(s)-len(q)] + "-" + q
		}
	}
	return s
}

func normalizeOKXInterval(interval string) string {
	i := strings.ToLower(strings.TrimSpace(interval))
	switch i {
	case "1s", "1m", "3m", "5m", "15m", "30m":
		return i
	case "1h", "2h", "4h", "6h", "12h":
		return strings.ToUpper(i)
	case "1d", "3d":
		return strings.ToUpper(i)
	case "1w":
		return "1W"
	case "1mo", "1mth":
		return "1M"
	default:
		return i
	}
}

func getCandleDurationMs(interval string) int64 {
	switch interval {
	case "1s", "1m":
		return 60 * 1000
	case "3m":
		return 3 * 60 * 1000
	case "5m":
		return 5 * 60 * 1000
	case "15m":
		return 15 * 60 * 1000
	case "30m":
		return 30 * 60 * 1000
	case "1H":
		return 60 * 60 * 1000
	case "2H":
		return 2 * 60 * 60 * 1000
	case "4H":
		return 4 * 60 * 60 * 1000
	case "6H":
		return 6 * 60 * 60 * 1000
	case "12H":
		return 12 * 60 * 60 * 1000
	case "1D":
		return 24 * 60 * 60 * 1000
	case "1W":
		return 7 * 24 * 60 * 60 * 1000
	case "1M":
		return 30 * 24 * 60 * 60 * 1000
	default:
		return 60 * 1000
	}
}

func (sws *OkxSpotWebSocket) keepAlive(ctx context.Context) {
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sws.pingStop:
			return
		case <-ticker.C:
			if sws.conn != nil {
				_ = sws.conn.WriteMessage(websocket.TextMessage, []byte("ping"))
			}
		}
	}
}
