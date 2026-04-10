package oauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"codexass/internal/config"
)

type CallbackServer struct {
	server        *http.Server
	callbackCh    chan string
	errorCh       chan error
	expectedState string
}

func NewCallbackServer(state string) *CallbackServer {
	cb := &CallbackServer{
		callbackCh:    make(chan string, 1),
		errorCh:       make(chan error, 1),
		expectedState: state,
	}
	mux := http.NewServeMux()
	mux.HandleFunc(config.CallbackPath, cb.handle)
	cb.server = &http.Server{Addr: fmt.Sprintf("%s:%d", config.CallbackHost, config.CallbackPort), Handler: mux}
	return cb
}

func (c *CallbackServer) handle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	state := query.Get("state")
	code := query.Get("code")
	if state != c.expectedState || code == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(loadTemplate("error.html")))
		select {
		case c.errorCh <- fmt.Errorf("invalid or expired authentication state"):
		default:
		}
		return
	}
	full := &url.URL{Scheme: "http", Host: r.Host, Path: r.URL.Path, RawQuery: r.URL.RawQuery}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(loadTemplate("success.html")))
	select {
	case c.callbackCh <- full.String():
	default:
	}
}

func (c *CallbackServer) Start() error {
	go func() {
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			select {
			case c.errorCh <- err:
			default:
			}
		}
	}()
	time.Sleep(150 * time.Millisecond)
	return nil
}

func (c *CallbackServer) Wait(timeout time.Duration) (string, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case callback := <-c.callbackCh:
		return callback, nil
	case err := <-c.errorCh:
		return "", err
	case <-timer.C:
		return "", fmt.Errorf("timed out waiting for OAuth callback")
	}
}

func (c *CallbackServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = c.server.Shutdown(ctx)
}
