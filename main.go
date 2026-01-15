// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/lrstanley/clix/v2"
	"github.com/lrstanley/outline-export/internal/api"
)

var (
	version = "master"
	commit  = "latest"
	date    = "-"
	cli     = clix.NewWithDefaults(
		clix.WithAppInfo[Flags](clix.AppInfo{
			Version: version,
			Commit:  commit,
			Date:    date,
		}),
		clix.WithKongOptions[Flags](kong.Vars{
			"HTTP_TIMEOUT": api.DefaultHTTPTimeout.Round(time.Second).String(),
		}),
	)
)

type Flags struct {
	URL                string        `name:"url" env:"URL" required:"" help:"URL of the Outline server"`
	Token              string        `name:"token" env:"TOKEN" required:"" help:"Token for the Outline server"`
	Format             string        `name:"format" env:"FORMAT" required:"" enum:"markdown,html,json" help:"Format of the export"`
	ExcludeAttachments bool          `name:"exclude-attachments" env:"EXCLUDE_ATTACHMENTS" help:"Exclude attachments from the export"`
	ExcludePrivate     bool          `name:"exclude-private" env:"EXCLUDE_PRIVATE" help:"Exclude private collections from the export"`
	Extract            bool          `name:"extract" env:"EXTRACT" help:"Extract the export into the target directory"`
	ExportPath         string        `name:"export-path" env:"EXPORT_PATH" required:"" help:"Path to export the file to. If extract is enabled, this will be the directory to extract the export to."`
	Filters            []string      `name:"filters" env:"FILTERS" help:"Filters the export to only include certain files (when using --extract). This is a glob pattern, and it matches the files/folders inside of the export zip, not necessarily collections/document exact names."`
	HTTPTimeout        time.Duration `name:"http-timeout" env:"HTTP_TIMEOUT" default:"${HTTP_TIMEOUT}" help:"Timeout for HTTP requests to the Outline server"`
}

func main() {
	ctx := context.Background()
	logger := cli.GetLogger()

	client, err := api.NewClient(&api.Config{
		BaseURL:     cli.Flags.URL,
		Token:       cli.Flags.Token,
		Logger:      logger,
		HTTPTimeout: cli.Flags.HTTPTimeout,
	})
	if err != nil {
		logger.Error("failed to create client", "error", err)
		os.Exit(1)
	}

	var format api.ExportFormat
	switch cli.Flags.Format {
	case "markdown":
		format = api.ExportFormatMarkdown
	case "html":
		format = api.ExportFormatHTML
	case "json":
		format = api.ExportFormatJSON
	default:
		logger.Error("invalid format", "format", cli.Flags.Format)
		os.Exit(1)
	}

	var operation *api.FileOperation

	for op, err := range client.ListFileOperations(ctx) {
		if err != nil {
			logger.Error("failed to list file operations", "error", err)
			os.Exit(1)
		}

		if op.Type != api.FileOperationTypeExport || op.Format != format || time.Since(op.CreatedAt) > 1*time.Hour {
			continue
		}

		if op.State == api.FileOperationStateComplete || op.State == api.FileOperationStateCreating || op.State == api.FileOperationStateUploading {
			logger.Info("found existing export", "id", op.ID, "name", op.Name)
			operation = op
			break
		}
	}

	if operation == nil {
		operation, err = client.GenerateExport(ctx, format, !cli.Flags.ExcludeAttachments, !cli.Flags.ExcludePrivate)
		if err != nil {
			logger.Error("failed to generate export", "error", err)
			os.Exit(1)
		}
	}

	// Wait for the operation to complete.
	tctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	operation, err = client.WaitForFileOperation(tctx, operation.ID)
	if err != nil {
		logger.Error("failed to wait for file operation", "error", err)
		os.Exit(1)
	}

	err = downloadExport(ctx, client, operation)
	if err != nil {
		logger.Error("failed to download export", "error", err)
		os.Exit(1)
	}
	logger.Info("export downloaded")

	// Delete all recently created exports, within the last 1 hour, that match our format.
	for op, err := range client.ListFileOperations(ctx) {
		if err != nil {
			logger.Error("failed to list file operations", "error", err)
			os.Exit(1)
		}

		if op.Type != api.FileOperationTypeExport || op.Format != format || op.State == api.FileOperationStateError || time.Since(op.CreatedAt) > 1*time.Hour {
			continue
		}

		err = client.DeleteFileOperation(ctx, op.ID)
		if err != nil {
			logger.Error(
				"failed to delete export",
				"id", op.ID,
				"name", op.Name,
				"error", err,
			)
			continue
		}

		logger.Info("deleted export", "id", op.ID, "name", op.Name)
	}
}

var (
	reInvalid     = regexp.MustCompile(`[^a-zA-Z0-9_.~\[\]()& -]+`)
	reCleanDashes = regexp.MustCompile(`-+`)
)

func downloadExport(ctx context.Context, client *api.Client, operation *api.FileOperation) error {
	// Download the export.
	reader, err := client.DownloadFileExport(ctx, operation.ID)
	if reader != nil {
		defer reader.Close()
	}

	if err != nil {
		return fmt.Errorf("failed to download export: %w", err)
	}

	if !cli.Flags.Extract {
		err = os.MkdirAll(path.Dir(cli.Flags.ExportPath), 0o700)
		if err != nil {
			return fmt.Errorf("failed to create export directory %q: %w", cli.Flags.ExportPath, err)
		}

		f, err := os.OpenFile(cli.Flags.ExportPath, os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("failed to initialize export file %q: %w", cli.Flags.ExportPath, err)
		}

		_, err = io.Copy(f, reader)
		if err != nil {
			return fmt.Errorf("failed to copy export to file %q: %w", cli.Flags.ExportPath, err)
		}

		slog.InfoContext(ctx, "export file written", "file", cli.Flags.ExportPath)
		return nil
	}

	err = os.MkdirAll(cli.Flags.ExportPath, 0o700)
	if err != nil {
		return fmt.Errorf("failed to create export directory %q: %w", cli.Flags.ExportPath, err)
	}

	tmp, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("outline-export-%s-*.zip", operation.ID))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmp.Name())

	length, err := io.Copy(tmp, reader)
	if err != nil {
		return fmt.Errorf("failed to stream export to temporary file: %w", err)
	}

	zr, err := zip.NewReader(tmp, length)
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	for _, f := range zr.File {

		// Sanitize parts.
		parts := strings.Split(f.Name, "/")
		for i := range parts {
			// URL decode the part.
			parts[i], err = url.QueryUnescape(parts[i])
			if err != nil {
				return fmt.Errorf("failed to unescape path part %q: %w", parts[i], err)
			}

			// Replace any potentially unsupported characters with a dash.
			parts[i] = reInvalid.ReplaceAllString(parts[i], "-")
			// Clean up any double dashes.
			parts[i] = reCleanDashes.ReplaceAllString(parts[i], "-")
			// Remove any leading/trailing dashes.
			parts[i] = strings.Trim(parts[i], "-")
		}

		// Join the parts back together.
		name := filepath.Join(parts...)

		if len(cli.Flags.Filters) > 0 {
			var matched bool
			for _, filter := range cli.Flags.Filters {
				matched, err = filepath.Match(filter, name)
				if err != nil {
					return fmt.Errorf("invalid filter pattern %q: %w", filter, err)
				}
				if matched {
					break
				}
			}

			if !matched {
				slog.WarnContext(ctx, "skipping file/folder (does not match filter)", "path", name)
				continue
			}
		}

		inf, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %q: %w", name, err)
		}

		if f.FileInfo().IsDir() {
			slog.InfoContext(ctx, "creating directory", "path", name)

			err = os.MkdirAll(path.Join(cli.Flags.ExportPath, name), 0o700)
			if err != nil {
				_ = inf.Close()
				return fmt.Errorf("failed to create directory %q: %w", name, err)
			}

			_ = inf.Close()
			continue
		}

		slog.InfoContext(ctx, "creating file", "path", name)
		outf, err := os.OpenFile(path.Join(cli.Flags.ExportPath, name), os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			_ = inf.Close()
			return fmt.Errorf("failed to create file %q: %w", name, err)
		}

		_, err = io.Copy(outf, inf)
		if err != nil {
			_ = inf.Close()
			_ = outf.Close()
			return fmt.Errorf("failed to copy file %q: %w", name, err)
		}

		_ = inf.Close()
		_ = outf.Close()
	}
	return nil
}
