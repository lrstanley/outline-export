// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/lrstanley/clix"
	"github.com/lrstanley/outline-export/internal/api"
)

var (
	version = "master"
	commit  = "latest"
	date    = "-"

	cli = &clix.CLI[Flags]{
		Links: clix.GithubLinks("github.com/lrstanley/outline-export", "master", "https://liam.sh"),
		VersionInfo: &clix.VersionInfo[Flags]{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	}
)

type Flags struct {
	URL                string   `long:"url" env:"URL" required:"true" description:"URL of the Outline server"`
	Token              string   `long:"token" env:"TOKEN" required:"true" description:"Token for the Outline server"`
	Format             string   `long:"format" env:"FORMAT" required:"true" choice:"markdown" choice:"html" choice:"json" description:"Format of the export"`
	ExcludeAttachments bool     `long:"exclude-attachments" env:"EXCLUDE_ATTACHMENTS" description:"Exclude attachments from the export"`
	Extract            bool     `long:"extract" env:"EXTRACT" description:"Extract the export into the target directory"`
	ExportPath         string   `long:"export-path" env:"EXPORT_PATH" required:"true" description:"Path to export the file to. If extract is enabled, this will be the directory to extract the export to."`
	Filters            []string `long:"filter" env:"FILTER" env-delim:"," description:"Filters the export to only include certain files (when using --extract). This is a glob pattern, and it matches the files/folders inside of the export zip, not necessarily collections/document exact names. Can be specified multiple times."`
}

func main() {
	cli.LoggerConfig.Pretty = true
	cli.Parse()

	ctx := log.NewContext(context.Background(), cli.Logger)

	client, err := api.NewClient(&api.Config{
		BaseURL: cli.Flags.URL,
		Token:   cli.Flags.Token,
	})
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("failed to create client")
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
		log.FromContext(ctx).WithField("format", cli.Flags.Format).Fatal("invalid format")
	}

	var operation *api.FileOperation

	for op, err := range client.ListFileOperations(ctx) {
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("failed to list file operations")
		}

		if op.Type != api.FileOperationTypeExport || op.Format != format || time.Since(op.CreatedAt) > 1*time.Hour {
			continue
		}

		if op.State == api.FileOperationStateComplete || op.State == api.FileOperationStateCreating || op.State == api.FileOperationStateUploading {
			log.FromContext(ctx).WithFields(log.Fields{
				"id":   op.ID,
				"name": op.Name,
			}).Info("found existing export")

			operation = op
			break
		}
	}

	if operation == nil {
		operation, err = client.GenerateExport(ctx, format, !cli.Flags.ExcludeAttachments)
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("failed to generate export")
		}
	}

	// Wait for the operation to complete.
	tctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	operation, err = client.WaitForFileOperation(tctx, operation.ID)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("failed to wait for file operation")
	}

	err = downloadExport(ctx, client, operation)
	if err != nil {
		log.FromContext(ctx).WithError(err).Fatal("failed to download export")
	}
	log.FromContext(ctx).Info("export downloaded")

	// Delete all recently created exports, within the last 1 hour, that match our format.
	for op, err := range client.ListFileOperations(ctx) {
		if err != nil {
			log.FromContext(ctx).WithError(err).Fatal("failed to list file operations")
		}

		if op.Type != api.FileOperationTypeExport || op.Format != format || time.Since(op.CreatedAt) > 1*time.Hour {
			continue
		}

		dctx := log.NewContext(ctx, log.FromContext(ctx).WithFields(log.Fields{
			"id":   op.ID,
			"name": op.Name,
		}))

		err = client.DeleteFileOperation(dctx, op.ID)
		if err != nil {
			log.FromContext(dctx).WithError(err).Fatal("failed to delete export")
		}

		log.FromContext(dctx).Info("deleted export")
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

		ctx := log.NewContext(ctx, log.FromContext(ctx).WithField("file", cli.Flags.ExportPath))

		f, err := os.OpenFile(cli.Flags.ExportPath, os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("failed to initialize export file %q: %w", cli.Flags.ExportPath, err)
		}

		_, err = io.Copy(f, reader)
		if err != nil {
			return fmt.Errorf("failed to copy export to file %q: %w", cli.Flags.ExportPath, err)
		}

		log.FromContext(ctx).Info("export file written")
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

		fctx := log.NewContext(ctx, log.FromContext(ctx).WithField("path", name))

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
				log.FromContext(fctx).Warn("skipping file/folder (does not match filter)")
				continue
			}
		}

		inf, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %q: %w", name, err)
		}

		if f.FileInfo().IsDir() {
			log.FromContext(fctx).Info("creating directory")

			err = os.MkdirAll(path.Join(cli.Flags.ExportPath, name), 0o700)
			if err != nil {
				_ = inf.Close()
				return fmt.Errorf("failed to create directory %q: %w", name, err)
			}

			_ = inf.Close()
			continue
		}

		log.FromContext(fctx).Info("creating file")
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
