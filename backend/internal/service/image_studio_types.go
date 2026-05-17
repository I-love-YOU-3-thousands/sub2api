package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	ImageStudioModeTextToImage  = "text_to_image"
	ImageStudioModeImageToImage = "image_to_image"

	ImageStudioTaskStatusPending   = "pending"
	ImageStudioTaskStatusRunning   = "running"
	ImageStudioTaskStatusSucceeded = "succeeded"
	ImageStudioTaskStatusFailed    = "failed"
	ImageStudioTaskStatusCanceled  = "canceled"

	ImageStudioAssetKindInput  = "input"
	ImageStudioAssetKindOutput = "output"

	ImageStudioStorageDriverLocal = "local"
	ImageStudioStorageDriverS3    = "s3"
)

var ErrImageStudioNoPendingTask = errors.New("no pending image studio task")

type ImageGenerationTask struct {
	ID           int64
	TaskID       string
	UserID       int64
	APIKeyID     int64
	GroupID      *int64
	Mode         string
	Model        string
	Prompt       string
	Ratio        string
	Resolution   string
	Size         string
	Quality      string
	Count        int
	Status       string
	Progress     int
	ErrorMessage *string
	RequestMeta  json.RawMessage
	AccountID    *int64
	UsageMeta    json.RawMessage
	StartedAt    *time.Time
	FinishedAt   *time.Time
	DeletedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Assets       []ImageGenerationAsset
}

type ImageGenerationAsset struct {
	ID            int64
	TaskID        string
	UserID        int64
	Seq           int
	Kind          string
	StorageDriver string
	StorageKey    string
	MimeType      string
	Width         int
	Height        int
	SizeBytes     int64
	OriginalURL   *string
	RevisedPrompt *string
	Meta          json.RawMessage
	DeletedAt     *time.Time
	CreatedAt     time.Time
}

type ImageGenerationUpstreamLog struct {
	ID              int64
	TaskID          string
	UserID          int64
	AccountID       *int64
	Provider        string
	Endpoint        string
	StatusCode      *int
	DurationMs      int64
	RequestExcerpt  string
	ResponseExcerpt string
	ErrorMessage    *string
	Meta            json.RawMessage
	CreatedAt       time.Time
}

type ImageGenerationRepository interface {
	CreateTask(ctx context.Context, task *ImageGenerationTask) error
	GetTaskByTaskID(ctx context.Context, userID int64, taskID string, includeDeleted bool) (*ImageGenerationTask, error)
	ListTasksByUserID(ctx context.Context, userID int64, params pagination.PaginationParams) ([]ImageGenerationTask, *pagination.PaginationResult, error)
	SoftDeleteTask(ctx context.Context, userID int64, taskID string) error
	AddAsset(ctx context.Context, asset *ImageGenerationAsset) error
	ListAssetsByTaskID(ctx context.Context, taskID string) ([]ImageGenerationAsset, error)
	GetAssetByIDForUser(ctx context.Context, userID, assetID int64) (*ImageGenerationAsset, error)
	ClaimNextPendingTask(ctx context.Context, staleAfter time.Duration) (*ImageGenerationTask, error)
	RecoverRunningTasks(ctx context.Context, staleAfter time.Duration) (int64, error)
	UpdateTaskProgress(ctx context.Context, taskID string, progress int, usageMeta json.RawMessage) error
	UpdateTaskProgressOnly(ctx context.Context, taskID string, progress int) error
	SetTaskSucceeded(ctx context.Context, taskID string, usageMeta json.RawMessage) error
	SetTaskFailed(ctx context.Context, taskID string, message string, usageMeta json.RawMessage) error
	AppendUpstreamLog(ctx context.Context, log *ImageGenerationUpstreamLog) error
}

type ImageStudioStoreInput struct {
	TaskID        string
	UserID        int64
	Seq           int
	Kind          string
	FileName      string
	MimeType      string
	Data          []byte
	OriginalURL   *string
	RevisedPrompt *string
	Meta          json.RawMessage
}

type ImageStudioStoredAsset struct {
	StorageDriver string
	StorageKey    string
	MimeType      string
	Width         int
	Height        int
	SizeBytes     int64
}

type ImageStudioStorage interface {
	Save(ctx context.Context, input ImageStudioStoreInput) (*ImageStudioStoredAsset, error)
	Open(ctx context.Context, asset *ImageGenerationAsset) (io.ReadCloser, string, int64, error)
}

type ImageStudioKeySummary struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	GroupID   *int64 `json:"group_id,omitempty"`
	GroupName string `json:"group_name,omitempty"`
	Platform  string `json:"platform"`
	Status    string `json:"status"`
}
