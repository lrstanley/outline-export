// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// prepareRequest prepares a request for the given client, method, path, params, and body.
func prepareRequest(
	ctx context.Context,
	client *Client,
	method,
	path string,
	params map[string]string,
	body map[string]any,
) (*http.Request, error) {
	var buf bytes.Buffer

	if body != nil {
		enc := json.NewEncoder(&buf)
		if err := enc.Encode(body); err != nil {
			return nil, fmt.Errorf("failed to encode body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, client.Config.BaseURL+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Authorization", "Bearer "+client.Config.Token)
	req.Header.Set("User-Agent", "outline-export")

	if params != nil {
		query := req.URL.Query()
		for k, v := range params {
			query.Set(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	return req, nil
}

// request is a generic function that makes an HTTP request to the given path, with
// the given method, params, and body. If the type of T is a string, the body will be
// read and returned as a string, otherwise [request] will attempt to parse the body
// as JSON.
func request[T any](
	ctx context.Context,
	client *Client,
	method,
	path string,
	params map[string]string,
	body map[string]any,
) (T, error) {
	var result T

	req, err := prepareRequest(ctx, client, method, path, params, body)
	if err != nil {
		return result, err
	}

	logger := slog.With(
		"method", req.Method,
		"url", req.URL.String(),
	)

	logger.DebugContext(ctx, "sending request")
	start := time.Now()
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close() //nolint:errcheck

	logger = logger.With(
		"status", resp.Status,
		"duration", time.Since(start).Round(time.Millisecond),
	)

	if resp.StatusCode >= 299 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return result, fmt.Errorf("failed to read body: %w", err)
		}
		logger.ErrorContext(ctx, "request failed", "body", string(body))

		return result, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}
	logger.DebugContext(ctx, "request completed")

	// Decode and wrap in generics. if type of T is string, return the body as a string.
	if _, ok := any(result).(string); ok {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return result, err
		}
		result = any(string(body)).(T)
	} else if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}

func requestStream(
	ctx context.Context,
	client *Client,
	method,
	path string,
	params map[string]string,
	body map[string]any,
) (io.ReadCloser, error) {
	req, err := prepareRequest(ctx, client, method, path, params, body)
	if err != nil {
		return nil, err
	}

	logger := slog.With(
		"method", req.Method,
		"url", req.URL.String(),
	)

	logger.DebugContext(ctx, "sending request")
	start := time.Now()

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	logger = logger.With(
		"status", resp.Status,
		"duration", time.Since(start).Round(time.Millisecond),
	)

	if resp.StatusCode >= 299 {
		defer resp.Body.Close() //nolint:errcheck
		logger.ErrorContext(ctx, "request failed")
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	logger.DebugContext(ctx, "request completed")
	return resp.Body, nil
}
