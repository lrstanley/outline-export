// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"encoding/json"
	"time"
)

const (
	ExportFormatMarkdown ExportFormat = "outline-markdown"
	ExportFormatHTML     ExportFormat = "html"
	ExportFormatJSON     ExportFormat = "json"
)

type ExportFormat string

const (
	FileOperationStateComplete  FileOperationState = "complete"
	FileOperationStateCreating  FileOperationState = "creating"
	FileOperationStateError     FileOperationState = "error"
	FileOperationStateExpired   FileOperationState = "expired"
	FileOperationStateUploading FileOperationState = "uploading"
)

type FileOperationState string

const (
	FileOperationTypeExport FileOperationType = "export"
	FileOperationTypeImport FileOperationType = "import"
)

type FileOperationType string

type FileOperation struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Type      FileOperationType   `json:"type"`
	Format    ExportFormat        `json:"format"`
	State     FileOperationState  `json:"state"`
	Error     *FileOperationError `json:"error"`
	CreatedAt time.Time           `json:"createdAt"`
	UpdatedAt time.Time           `json:"updatedAt"`
}

type FileOperationError struct {
	Data    map[string]any `json:"data,omitempty"`
	Error   string         `json:"error,omitempty"`
	Message string         `json:"message,omitempty"`
	Ok      bool           `json:"ok,omitempty"`
	Status  float32        `json:"status,omitempty"`
}

func (e *FileOperationError) UnmarshalJSON(data []byte) error {
	// Error field may be a string, or an object.
	var serr struct {
		Data    map[string]any `json:"data,omitempty"`
		Error   string         `json:"error,omitempty"`
		Message string         `json:"message,omitempty"`
		Ok      bool           `json:"ok,omitempty"`
		Status  float32        `json:"status,omitempty"`
	}
	if err := json.Unmarshal(data, &serr); err == nil {
		e.Data = serr.Data
		e.Error = serr.Error
		e.Message = serr.Message
		e.Ok = serr.Ok
		e.Status = serr.Status
		return nil
	}
	err := json.Unmarshal(data, &e.Error)
	if err == nil {
		e.Message = e.Error
		e.Ok = false
		return nil
	}
	return err
}

type Pagination struct {
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Total    int    `json:"total"`
	NextPath string `json:"nextPath"`
}

type Abilities struct {
	Read   bool `json:"read"`
	Delete bool `json:"delete"`
}

type Policies struct {
	ID        string    `json:"id"`
	Abilities Abilities `json:"abilities"`
}
