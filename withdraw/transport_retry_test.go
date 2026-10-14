package withdraw

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gorilla/websocket"
)

func TestHTTPTimeoutRetry(t *testing.T) {
	for _, tc := range []struct {
		name       string
		retries    int
		deadline   time.Duration
		wantCalls  int32
		wantCancel bool
	}{
		{"transport timeout recovers", 1, 2 * time.Second, 2, false},
		{"transport timeout exhausts budget", 0, 2 * time.Second, 1, false},
		{"operation deadline is terminal", 1, 30 * time.Millisecond, 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req rpcRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Error(err)
					return
				}
				if calls.Add(1) == 1 {
					<-r.Context().Done()
					return
				}
				if err := json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": successfulReceipt()}); err != nil {
					t.Error(err)
				}
			}))
			t.Cleanup(server.Close)
			client, err := rpc.DialOptions(context.Background(), server.URL, rpc.WithHTTPClient(&http.Client{Timeout: 100 * time.Millisecond}))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(client.Close)
			ctx, cancel := context.WithTimeout(context.Background(), tc.deadline)
			defer cancel()
			config := &RetryConfig{MaxRetries: tc.retries, Delay: time.Millisecond}
			var firstErr error
			err = config.Do(ctx, "receipt", func() error {
				_, err := ethclient.NewClient(client).TransactionReceipt(ctx, common.Hash{})
				if firstErr == nil {
					firstErr = err
				}
				return err
			})
			var netErr net.Error
			if !errors.Is(firstErr, context.DeadlineExceeded) || !errors.As(firstErr, &netErr) || !netErr.Timeout() {
				t.Fatalf("expected wrapped transport/context timeout, got %v", firstErr)
			}
			if calls.Load() != tc.wantCalls || tc.wantCancel != (ctx.Err() != nil) || (err == nil) != (tc.wantCalls == 2) {
				t.Fatalf("error %v, context %v, calls %d", err, ctx.Err(), calls.Load())
			}
		})
	}
}

func TestWebsocketDisconnectRetry(t *testing.T) {
	for _, code := range []int{
		websocket.CloseAbnormalClosure, websocket.CloseGoingAway, websocket.CloseServiceRestart, websocket.CloseTryAgainLater,
		websocket.CloseNormalClosure, websocket.CloseProtocolError, websocket.ClosePolicyViolation, websocket.CloseUnsupportedData, websocket.CloseMessageTooBig, websocket.CloseInternalServerErr,
	} {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			t.Parallel()
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
				if err != nil {
					t.Error(err)
					return
				}
				defer conn.Close()
				var req rpcRequest
				if err := conn.ReadJSON(&req); err != nil {
					t.Error(err)
					return
				}
				if calls.Add(1) == 1 {
					if code != websocket.CloseAbnormalClosure {
						if err := conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, "disconnect"), time.Now().Add(time.Second)); err != nil {
							t.Error(err)
						}
					}
					return
				}
				if err := conn.WriteJSON(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": successfulReceipt()}); err != nil {
					t.Error(err)
				}
			}))
			t.Cleanup(server.Close)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			client, err := rpc.DialContext(ctx, "ws"+strings.TrimPrefix(server.URL, "http"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(client.Close)
			config := &RetryConfig{MaxRetries: 1, Delay: time.Millisecond}
			err = waitForConfirmation(ctx, ethclient.NewClient(client), common.Hash{}, "proof", config)
			transient := code == websocket.CloseAbnormalClosure || code == websocket.CloseGoingAway || code == websocket.CloseServiceRestart || code == websocket.CloseTryAgainLater
			if transient {
				if err != nil || calls.Load() != 2 {
					t.Fatalf("close %d did not recover: %v after %d calls", code, err, calls.Load())
				}
			} else {
				var closeErr *websocket.CloseError
				if !errors.As(err, &closeErr) || closeErr.Code != code || calls.Load() != 1 {
					t.Fatalf("close %d must be terminal: %v after %d calls", code, err, calls.Load())
				}
			}
		})
	}
}
