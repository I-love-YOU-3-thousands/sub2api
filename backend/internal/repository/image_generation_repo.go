package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type imageGenerationRepository struct {
	sql sqlExecutor
}

func NewImageGenerationRepository(_ *dbent.Client, sqlDB *sql.DB) service.ImageGenerationRepository {
	return &imageGenerationRepository{sql: sqlDB}
}

func (r *imageGenerationRepository) CreateTask(ctx context.Context, task *service.ImageGenerationTask) error {
	if task == nil {
		return errors.New("image generation task is nil")
	}
	query := `
INSERT INTO image_generation_tasks (
	task_id, user_id, api_key_id, group_id, mode, model, prompt, ratio, resolution, size, quality, count,
	status, progress, request_meta, usage_meta
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
	$13, $14, $15, $16
)
RETURNING id, created_at, updated_at`
	return scanSingleRow(ctx, r.sql, query, []any{
		task.TaskID,
		task.UserID,
		task.APIKeyID,
		task.GroupID,
		task.Mode,
		task.Model,
		task.Prompt,
		task.Ratio,
		task.Resolution,
		task.Size,
		task.Quality,
		task.Count,
		task.Status,
		task.Progress,
		jsonOrEmpty(task.RequestMeta),
		jsonOrEmpty(task.UsageMeta),
	}, &task.ID, &task.CreatedAt, &task.UpdatedAt)
}

func (r *imageGenerationRepository) GetTaskByTaskID(ctx context.Context, userID int64, taskID string, includeDeleted bool) (*service.ImageGenerationTask, error) {
	where := "task_id = $1 AND user_id = $2"
	if !includeDeleted {
		where += " AND deleted_at IS NULL"
	}
	task, err := r.scanTask(ctx, "SELECT "+imageGenerationTaskColumns+" FROM image_generation_tasks WHERE "+where, taskID, userID)
	if err != nil {
		return nil, err
	}
	assets, err := r.ListAssetsByTaskID(ctx, task.TaskID)
	if err != nil {
		return nil, err
	}
	task.Assets = assets
	return task, nil
}

func (r *imageGenerationRepository) ListTasksByUserID(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.ImageGenerationTask, *pagination.PaginationResult, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	var total int64
	if err := scanSingleRow(ctx, r.sql, `SELECT COUNT(*) FROM image_generation_tasks WHERE user_id = $1 AND deleted_at IS NULL`, []any{userID}, &total); err != nil {
		return nil, nil, err
	}
	rows, err := r.sql.QueryContext(ctx, "SELECT "+imageGenerationTaskColumns+` FROM image_generation_tasks
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`, userID, params.PageSize, params.Offset())
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	tasks := make([]service.ImageGenerationTask, 0, params.PageSize)
	for rows.Next() {
		task, err := scanImageGenerationTask(rows)
		if err != nil {
			return nil, nil, err
		}
		tasks = append(tasks, *task)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	for i := range tasks {
		assets, err := r.ListAssetsByTaskID(ctx, tasks[i].TaskID)
		if err != nil {
			return nil, nil, err
		}
		tasks[i].Assets = assets
	}
	pages := int(math.Ceil(float64(total) / float64(params.PageSize)))
	if pages < 1 {
		pages = 1
	}
	return tasks, &pagination.PaginationResult{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Pages:    pages,
	}, nil
}

func (r *imageGenerationRepository) SoftDeleteTask(ctx context.Context, userID int64, taskID string) error {
	res, err := r.sql.ExecContext(ctx, `UPDATE image_generation_tasks
SET deleted_at = NOW(), updated_at = NOW()
WHERE task_id = $1 AND user_id = $2 AND deleted_at IS NULL`, taskID, userID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	_, _ = r.sql.ExecContext(ctx, `UPDATE image_generation_assets SET deleted_at = NOW() WHERE task_id = $1 AND user_id = $2 AND deleted_at IS NULL`, taskID, userID)
	return nil
}

func (r *imageGenerationRepository) AddAsset(ctx context.Context, asset *service.ImageGenerationAsset) error {
	if asset == nil {
		return errors.New("image generation asset is nil")
	}
	query := `
INSERT INTO image_generation_assets (
	task_id, user_id, seq, kind, storage_driver, storage_key, mime_type, width, height, size_bytes,
	original_url, revised_prompt, meta
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
	$11, $12, $13
)
RETURNING id, created_at`
	return scanSingleRow(ctx, r.sql, query, []any{
		asset.TaskID,
		asset.UserID,
		asset.Seq,
		asset.Kind,
		asset.StorageDriver,
		asset.StorageKey,
		asset.MimeType,
		asset.Width,
		asset.Height,
		asset.SizeBytes,
		asset.OriginalURL,
		asset.RevisedPrompt,
		jsonOrEmpty(asset.Meta),
	}, &asset.ID, &asset.CreatedAt)
}

func (r *imageGenerationRepository) ListAssetsByTaskID(ctx context.Context, taskID string) ([]service.ImageGenerationAsset, error) {
	rows, err := r.sql.QueryContext(ctx, `SELECT `+imageGenerationAssetColumns+`
FROM image_generation_assets
WHERE task_id = $1 AND deleted_at IS NULL
ORDER BY kind ASC, seq ASC, id ASC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	assets := []service.ImageGenerationAsset{}
	for rows.Next() {
		asset, err := scanImageGenerationAsset(rows)
		if err != nil {
			return nil, err
		}
		assets = append(assets, *asset)
	}
	return assets, rows.Err()
}

func (r *imageGenerationRepository) GetAssetByIDForUser(ctx context.Context, userID, assetID int64) (*service.ImageGenerationAsset, error) {
	return r.scanAsset(ctx, `SELECT `+imageGenerationAssetColumns+`
FROM image_generation_assets
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, assetID, userID)
}

func (r *imageGenerationRepository) ClaimNextPendingTask(ctx context.Context, staleAfter time.Duration) (*service.ImageGenerationTask, error) {
	_, _ = r.sql.ExecContext(ctx, `UPDATE image_generation_tasks
SET status = $1, progress = 0, error_message = NULL, started_at = NULL, updated_at = NOW()
WHERE status = $2 AND started_at < NOW() - ($3::bigint * INTERVAL '1 second') AND deleted_at IS NULL`,
		service.ImageStudioTaskStatusPending,
		service.ImageStudioTaskStatusRunning,
		int64(staleAfter.Seconds()),
	)
	query := `
WITH next_task AS (
	SELECT id
	FROM image_generation_tasks
	WHERE status = $1 AND deleted_at IS NULL
	ORDER BY created_at ASC
	LIMIT 1
	FOR UPDATE SKIP LOCKED
)
UPDATE image_generation_tasks t
SET status = $2, progress = 5, started_at = NOW(), updated_at = NOW()
FROM next_task
WHERE t.id = next_task.id
RETURNING ` + imageGenerationTaskColumnsWithAlias("t")
	task, err := r.scanTask(ctx, query, service.ImageStudioTaskStatusPending, service.ImageStudioTaskStatusRunning)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrImageStudioNoPendingTask
	}
	return task, err
}

func (r *imageGenerationRepository) RecoverRunningTasks(ctx context.Context, staleAfter time.Duration) (int64, error) {
	res, err := r.sql.ExecContext(ctx, `UPDATE image_generation_tasks
SET status = $1, progress = 0, error_message = NULL, started_at = NULL, updated_at = NOW()
WHERE status = $2 AND started_at < NOW() - ($3::bigint * INTERVAL '1 second') AND deleted_at IS NULL`,
		service.ImageStudioTaskStatusPending,
		service.ImageStudioTaskStatusRunning,
		int64(staleAfter.Seconds()),
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *imageGenerationRepository) UpdateTaskProgress(ctx context.Context, taskID string, progress int, usageMeta json.RawMessage) error {
	_, err := r.sql.ExecContext(ctx, `UPDATE image_generation_tasks
	SET progress = GREATEST(progress, $2), usage_meta = $3, updated_at = NOW()
	WHERE task_id = $1 AND deleted_at IS NULL`, taskID, progress, jsonOrEmpty(usageMeta))
	return err
}

func (r *imageGenerationRepository) UpdateTaskProgressOnly(ctx context.Context, taskID string, progress int) error {
	_, err := r.sql.ExecContext(ctx, `UPDATE image_generation_tasks
	SET progress = GREATEST(progress, $2), updated_at = NOW()
	WHERE task_id = $1 AND status = $3 AND deleted_at IS NULL`, taskID, progress, service.ImageStudioTaskStatusRunning)
	return err
}

func (r *imageGenerationRepository) SetTaskSucceeded(ctx context.Context, taskID string, usageMeta json.RawMessage) error {
	_, err := r.sql.ExecContext(ctx, `UPDATE image_generation_tasks
	SET status = $2, progress = 100, error_message = NULL, usage_meta = $3, finished_at = NOW(), updated_at = NOW()
WHERE task_id = $1 AND deleted_at IS NULL`, taskID, service.ImageStudioTaskStatusSucceeded, jsonOrEmpty(usageMeta))
	return err
}

func (r *imageGenerationRepository) SetTaskFailed(ctx context.Context, taskID string, message string, usageMeta json.RawMessage) error {
	_, err := r.sql.ExecContext(ctx, `UPDATE image_generation_tasks
SET status = $2, progress = CASE WHEN progress < 5 THEN 5 ELSE progress END, error_message = $3, usage_meta = $4, finished_at = NOW(), updated_at = NOW()
WHERE task_id = $1 AND deleted_at IS NULL`, taskID, service.ImageStudioTaskStatusFailed, message, jsonOrEmpty(usageMeta))
	return err
}

func (r *imageGenerationRepository) AppendUpstreamLog(ctx context.Context, log *service.ImageGenerationUpstreamLog) error {
	if log == nil {
		return nil
	}
	query := `
INSERT INTO image_generation_upstream_logs (
	task_id, user_id, account_id, provider, endpoint, status_code, duration_ms,
	request_excerpt, response_excerpt, error_message, meta
) VALUES (
	$1, $2, $3, $4, $5, $6, $7,
	$8, $9, $10, $11
) RETURNING id, created_at`
	return scanSingleRow(ctx, r.sql, query, []any{
		log.TaskID,
		log.UserID,
		log.AccountID,
		log.Provider,
		log.Endpoint,
		log.StatusCode,
		log.DurationMs,
		truncateImageStudioExcerpt(log.RequestExcerpt),
		truncateImageStudioExcerpt(log.ResponseExcerpt),
		log.ErrorMessage,
		jsonOrEmpty(log.Meta),
	}, &log.ID, &log.CreatedAt)
}

const imageGenerationTaskColumns = `id, task_id, user_id, api_key_id, group_id, mode, model, prompt, ratio, resolution, size, quality, count, status, progress, error_message, request_meta, account_id, usage_meta, started_at, finished_at, deleted_at, created_at, updated_at`
const imageGenerationAssetColumns = `id, task_id, user_id, seq, kind, storage_driver, storage_key, mime_type, width, height, size_bytes, original_url, revised_prompt, meta, deleted_at, created_at`

func imageGenerationTaskColumnsWithAlias(alias string) string {
	parts := strings.Split(imageGenerationTaskColumns, ", ")
	for i, part := range parts {
		parts[i] = alias + "." + part
	}
	return strings.Join(parts, ", ")
}

func (r *imageGenerationRepository) scanTask(ctx context.Context, query string, args ...any) (*service.ImageGenerationTask, error) {
	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, sql.ErrNoRows
	}
	task, err := scanImageGenerationTask(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return task, nil
}

func (r *imageGenerationRepository) scanAsset(ctx context.Context, query string, args ...any) (*service.ImageGenerationAsset, error) {
	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, sql.ErrNoRows
	}
	asset, err := scanImageGenerationAsset(rows)
	if err != nil {
		return nil, err
	}
	return asset, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanImageGenerationTask(row rowScanner) (*service.ImageGenerationTask, error) {
	var task service.ImageGenerationTask
	if err := row.Scan(
		&task.ID,
		&task.TaskID,
		&task.UserID,
		&task.APIKeyID,
		&task.GroupID,
		&task.Mode,
		&task.Model,
		&task.Prompt,
		&task.Ratio,
		&task.Resolution,
		&task.Size,
		&task.Quality,
		&task.Count,
		&task.Status,
		&task.Progress,
		&task.ErrorMessage,
		&task.RequestMeta,
		&task.AccountID,
		&task.UsageMeta,
		&task.StartedAt,
		&task.FinishedAt,
		&task.DeletedAt,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &task, nil
}

func scanImageGenerationAsset(row rowScanner) (*service.ImageGenerationAsset, error) {
	var asset service.ImageGenerationAsset
	if err := row.Scan(
		&asset.ID,
		&asset.TaskID,
		&asset.UserID,
		&asset.Seq,
		&asset.Kind,
		&asset.StorageDriver,
		&asset.StorageKey,
		&asset.MimeType,
		&asset.Width,
		&asset.Height,
		&asset.SizeBytes,
		&asset.OriginalURL,
		&asset.RevisedPrompt,
		&asset.Meta,
		&asset.DeletedAt,
		&asset.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &asset, nil
}

func jsonOrEmpty(raw json.RawMessage) any {
	if len(raw) == 0 {
		return []byte("{}")
	}
	return raw
}

func truncateImageStudioExcerpt(s string) string {
	const max = 4000
	if len(s) <= max {
		return s
	}
	return fmt.Sprintf("%s...", s[:max])
}
