// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBaseURL     = "https://app.getoutline.com"
	DefaultHTTPTimeout = 60 * time.Second
)

type Config struct {
	BaseURL     string
	Token       string
	Logger      *slog.Logger
	HTTPTimeout time.Duration
}

type Client struct {
	HTTPClient *http.Client
	Config     *Config
}

func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = &Config{}
	}

	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = DefaultHTTPTimeout
	}

	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	if config.BaseURL == "" {
		config.BaseURL = DefaultBaseURL
	}

	config.BaseURL = strings.TrimSuffix(strings.TrimSuffix(config.BaseURL, "/"), "/api") + "/api"

	if config.Token == "" {
		return nil, errors.New("token is required")
	}

	return &Client{
		HTTPClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
		Config: config,
	}, nil
}

// GenerateExport generates an export of all collections. Note that it will likely
// be pending once returned, and you should poll until it's ready. See
// [WaitForFileOperation] and [GenerateExportAndWait] for more information.
func (c *Client) GenerateExport(ctx context.Context, format ExportFormat, includeAttachments bool, includePrivate bool) (*FileOperation, error) {
	type Response struct {
		Data struct {
			FileOperation *FileOperation `json:"fileOperation"`
		} `json:"data"`
	}
	r, err := request[*Response](
		ctx, c, http.MethodPost,
		"/collections.export_all",
		nil,
		map[string]any{
			"format":             format,
			"includeAttachments": includeAttachments,
			"includePrivate":     includePrivate,
		},
	)
	if err != nil {
		return nil, err
	}
	return r.Data.FileOperation, nil
}

// GenerateExportAndWait generates an export of all collections and waits for it to complete.
func (c *Client) GenerateExportAndWait(ctx context.Context, format ExportFormat, includeAttachments bool, includePrivate bool) (*FileOperation, error) {
	op, err := c.GenerateExport(ctx, format, includeAttachments, includePrivate)
	if err != nil {
		return nil, err
	}
	return c.WaitForFileOperation(ctx, op.ID)
}

// DownloadFileExport downloads a file export.
func (c *Client) DownloadFileExport(ctx context.Context, id string) (io.ReadCloser, error) {
	return requestStream(
		ctx, c, http.MethodGet,
		"/fileOperations.redirect",
		map[string]string{"id": id},
		nil,
	)
}

// GetFileOperation fetches a specific file operation.
func (c *Client) GetFileOperation(ctx context.Context, id string) (*FileOperation, error) {
	type Response struct {
		Data *FileOperation `json:"data"`
	}
	r, err := request[*Response](
		ctx, c, http.MethodPost,
		"/fileOperations.info",
		nil,
		map[string]any{"id": id},
	)
	if err != nil {
		return nil, err
	}
	return r.Data, nil
}

// DeleteFileOperation deletes a file operation.
func (c *Client) DeleteFileOperation(ctx context.Context, id string) error {
	_, err := request[string](
		ctx, c, http.MethodPost,
		"/fileOperations.delete",
		nil,
		map[string]any{"id": id},
	)
	if err != nil {
		return err
	}
	return nil
}

// ListFileOperations lists all file operations.
func (c *Client) ListFileOperations(ctx context.Context) iter.Seq2[*FileOperation, error] {
	type Response struct {
		Pagination Pagination      `json:"pagination"`
		Data       []FileOperation `json:"data"`
		Policies   []Policies      `json:"policies"`
	}

	return func(yield func(*FileOperation, error) bool) {
		limit := 25
		offset := 0
		count := 0

		for {
			r, err := request[*Response](
				ctx, c, http.MethodPost,
				"/fileOperations.list",
				nil,
				map[string]any{
					"limit":  strconv.Itoa(limit),
					"offset": strconv.Itoa(offset),
					"type":   "export",
				},
			)
			if err != nil {
				yield(nil, err)
				return
			}

			for _, fileOperation := range r.Data {
				count++
				if !yield(&fileOperation, nil) {
					return
				}
			}

			if r.Pagination.Total <= count {
				return
			}

			offset += limit
			time.Sleep(250 * time.Millisecond)
		}
	}
}

// WaitForFileOperation waits for a file operation to complete. Use a context
// to cancel the operation if it takes too long.
func (c *Client) WaitForFileOperation(ctx context.Context, id string) (*FileOperation, error) {
	for {
		op, err := c.GetFileOperation(ctx, id)
		if err != nil {
			return nil, err
		}

		switch op.State {
		case FileOperationStateComplete:
			return op, nil
		case FileOperationStateError:
			return nil, fmt.Errorf("file operation failed: %s", op.Error.Message)
		case FileOperationStateExpired:
			return nil, errors.New("file operation expired")
		case FileOperationStateCreating, FileOperationStateUploading:
			slog.InfoContext(ctx, "waiting for file operation to complete", "state", op.State)
			time.Sleep(2 * time.Second)
			continue
		default:
			return nil, fmt.Errorf("unknown file operation state: %s", op.State)
		}
	}
}
