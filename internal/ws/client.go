package ws

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aaronfaby/kalshi-perp-cli/internal/auth"
	"nhooyr.io/websocket"
)

// Client streams margin WebSocket channels.
type Client struct {
	URL        string
	KeyID      string
	PrivateKey *rsa.PrivateKey
	Logf       func(format string, args ...any)
}

// Message is a generic WS envelope.
type Message map[string]any

// SubscribeParams configures channel subscriptions.
type SubscribeParams struct {
	Channels []string
	Tickers  []string
}

const (
	initialWSBackoff = time.Second
	maxWSBackoff     = 30 * time.Second
	wsPingInterval   = 15 * time.Second
	wsPingWait       = 10 * time.Second
)

func nextWSBackoff(cur time.Duration) time.Duration {
	next := cur * 2
	if next > maxWSBackoff {
		return maxWSBackoff
	}
	return next
}

// Run connects, subscribes, and writes messages to emit until ctx is done.
func (c *Client) Run(ctx context.Context, params SubscribeParams, emit func(Message) error) error {
	return c.runSession(ctx, params, emit, nil)
}

func (c *Client) runSession(ctx context.Context, params SubscribeParams, emit func(Message) error, onReady func()) error {
	headers, err := auth.Headers(c.KeyID, c.PrivateKey, http.MethodGet, wsSignPath(c.URL))
	if err != nil {
		return err
	}
	hdr := http.Header{}
	for k, v := range headers {
		hdr.Set(k, v)
	}

	conn, _, err := websocket.Dial(ctx, c.URL, &websocket.DialOptions{
		HTTPHeader: hdr,
	})
	if err != nil {
		return fmt.Errorf("ws dial: %w", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")

	// Kalshi sends Ping frames; nhooyr handles pong automatically for standard pings.
	conn.SetReadLimit(8 << 20)

	var msgID atomic.Int64
	if err := c.subscribe(ctx, conn, &msgID, params); err != nil {
		return err
	}
	if onReady != nil {
		onReady()
	}

	sessionCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	go pingWatchdog(sessionCtx, cancel, conn)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, data, err := conn.Read(sessionCtx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if cause := context.Cause(sessionCtx); cause != nil && !errors.Is(cause, context.Canceled) {
				return cause
			}
			if err == io.EOF {
				return fmt.Errorf("ws closed")
			}
			return fmt.Errorf("ws read: %w", err)
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			if err := emit(Message{"raw": string(data)}); err != nil {
				return err
			}
			continue
		}
		if err := emit(msg); err != nil {
			return err
		}
	}
}

func pingWatchdog(ctx context.Context, cancel context.CancelCauseFunc, conn *websocket.Conn) {
	t := time.NewTicker(wsPingInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			pingCtx, pingCancel := context.WithTimeout(ctx, wsPingWait)
			err := conn.Ping(pingCtx)
			pingCancel()
			if err != nil && ctx.Err() == nil {
				cancel(fmt.Errorf("ws idle ping: %w", err))
				return
			}
		}
	}
}

func (c *Client) subscribe(ctx context.Context, conn *websocket.Conn, msgID *atomic.Int64, params SubscribeParams) error {
	id := msgID.Add(1)
	b, err := BuildSubscribeMessage(id, params)
	if err != nil {
		return err
	}
	if c.Logf != nil {
		c.Logf("ws subscribe: %s", string(b))
	}
	return conn.Write(ctx, websocket.MessageText, b)
}

// BuildSubscribeMessage constructs the margin WS subscribe command JSON body.
func BuildSubscribeMessage(id int64, params SubscribeParams) ([]byte, error) {
	channels := params.Channels
	if len(channels) == 0 {
		channels = []string{"ticker"}
	}
	cmd := map[string]any{
		"id":  id,
		"cmd": "subscribe",
		"params": map[string]any{
			"channels": channels,
		},
	}
	if len(params.Tickers) > 0 {
		p := cmd["params"].(map[string]any)
		if len(params.Tickers) == 1 {
			p["market_ticker"] = params.Tickers[0]
		} else {
			p["market_tickers"] = params.Tickers
		}
	}
	return json.Marshal(cmd)
}

// RunWithReconnect loops Run with exponential backoff when reconnect is true.
func (c *Client) RunWithReconnect(ctx context.Context, params SubscribeParams, reconnect bool, emit func(Message) error) error {
	backoff := initialWSBackoff
	for {
		ready := false
		err := c.runSession(ctx, params, emit, func() { ready = true })
		if ready {
			backoff = initialWSBackoff
		}
		if err == nil || ctx.Err() != nil {
			return err
		}
		if !reconnect {
			return err
		}
		if c.Logf != nil {
			c.Logf("ws disconnected: %v; reconnecting in %s", err, backoff)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		backoff = nextWSBackoff(backoff)
	}
}

func wsSignPath(wsURL string) string {
	// wss://host/trade-api/ws/v2/margin -> /trade-api/ws/v2/margin
	// strip scheme and host
	u := wsURL
	if i := strings.Index(u, "://"); i >= 0 {
		u = u[i+3:]
	}
	if i := strings.Index(u, "/"); i >= 0 {
		return u[i:]
	}
	return "/trade-api/ws/v2/margin"
}
