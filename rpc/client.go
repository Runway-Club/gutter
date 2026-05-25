package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Call invokes a remote procedure with compile-time type safety. Req and Res
// are the shared structs; the route is derived from Req so it always matches
// the server's Handle registration. On js/wasm this issues a browser fetch via
// net/http's wasm transport; on the host it uses http.DefaultClient.
//
// A handler error on the server is returned here as a plain error carrying the
// server's message.
func Call[Req any, Res any](ctx context.Context, req Req) (Res, error) {
	var res Res

	payload, err := json.Marshal(req)
	if err != nil {
		return res, fmt.Errorf("gutter/rpc: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, Endpoint, bytes.NewReader(payload))
	if err != nil {
		return res, err
	}
	httpReq.Header.Set(procHeader, procName[Req]())
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return res, fmt.Errorf("gutter/rpc: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, fmt.Errorf("gutter/rpc: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var e errEnvelope
		if json.Unmarshal(body, &e) == nil && e.Error != "" {
			return res, errors.New(e.Error)
		}
		return res, fmt.Errorf("gutter/rpc: status %d", resp.StatusCode)
	}

	if err := json.Unmarshal(body, &res); err != nil {
		return res, fmt.Errorf("gutter/rpc: unmarshal response: %w", err)
	}
	return res, nil
}
