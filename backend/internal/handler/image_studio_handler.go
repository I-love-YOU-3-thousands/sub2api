package handler

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type ImageStudioHandler struct {
	service *service.ImageStudioService
}

func NewImageStudioHandler(svc *service.ImageStudioService) *ImageStudioHandler {
	return &ImageStudioHandler{service: svc}
}

func (h *ImageStudioHandler) ListKeys(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	keys, err := h.service.ListKeys(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, keys)
}

func (h *ImageStudioHandler) CreateTask(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		response.BadRequest(c, "invalid multipart form")
		return
	}
	apiKeyID, err := strconv.ParseInt(strings.TrimSpace(c.PostForm("api_key_id")), 10, 64)
	if err != nil || apiKeyID <= 0 {
		response.BadRequest(c, "api_key_id is required")
		return
	}
	count := 1
	if raw := strings.TrimSpace(c.PostForm("count")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			count = parsed
		}
	}
	input := service.ImageStudioCreateTaskInput{
		UserID:     subject.UserID,
		APIKeyID:   apiKeyID,
		Model:      c.PostForm("model"),
		Prompt:     c.PostForm("prompt"),
		Ratio:      c.PostForm("ratio"),
		Resolution: c.PostForm("resolution"),
		Quality:    c.PostForm("quality"),
		Count:      count,
		ClientIP:   c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	}
	if file, header, err := c.Request.FormFile("image"); err == nil {
		defer file.Close()
		data, readErr := io.ReadAll(io.LimitReader(file, service.ImageStudioMaxReferenceBytes()+1))
		if readErr != nil {
			response.BadRequest(c, "failed to read reference image")
			return
		}
		if int64(len(data)) > service.ImageStudioMaxReferenceBytes() {
			response.Error(c, http.StatusRequestEntityTooLarge, "reference image is too large")
			return
		}
		contentType := header.Header.Get("Content-Type")
		input.ReferenceImage = &service.ImageStudioUpload{
			FileName:    header.Filename,
			ContentType: contentType,
			Data:        data,
		}
	} else if err != nil && !errors.Is(err, http.ErrMissingFile) {
		response.BadRequest(c, "invalid reference image")
		return
	}
	task, err := h.service.CreateTask(c.Request.Context(), input)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Accepted(c, task)
}

func (h *ImageStudioHandler) GetTask(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	task, err := h.service.GetTask(c.Request.Context(), subject.UserID, c.Param("task_id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(c, "task not found")
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, task)
}

func (h *ImageStudioHandler) ListTasks(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	page, pageSize := response.ParsePagination(c)
	tasks, pag, err := h.service.ListTasks(c.Request.Context(), subject.UserID, pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    "created_at",
		SortOrder: pagination.SortOrderDesc,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, tasks, &response.PaginationResult{
		Total:    pag.Total,
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Pages:    pag.Pages,
	})
}

func (h *ImageStudioHandler) OptimizePrompt(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	var req struct {
		APIKeyID       int64  `json:"api_key_id"`
		Prompt         string `json:"prompt"`
		Ratio          string `json:"ratio"`
		Resolution     string `json:"resolution"`
		Quality        string `json:"quality"`
		PreviousPrompt string `json:"previous_prompt"`
		Variant        int    `json:"variant"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	result, err := h.service.OptimizePrompt(c.Request.Context(), service.ImageStudioOptimizePromptInput{
		UserID:         subject.UserID,
		APIKeyID:       req.APIKeyID,
		Prompt:         req.Prompt,
		Ratio:          req.Ratio,
		Resolution:     req.Resolution,
		Quality:        req.Quality,
		PreviousPrompt: req.PreviousPrompt,
		Variant:        req.Variant,
		ClientIP:       c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "prompt optimization timed out") {
			response.ErrorFrom(c, infraerrors.GatewayTimeout("PROMPT_OPTIMIZATION_TIMEOUT", "提示词优化超时，请稍后重试"))
			return
		}
		if errors.Is(err, context.Canceled) {
			response.ErrorFrom(c, infraerrors.ClientClosed("PROMPT_OPTIMIZATION_CANCELED", "提示词优化请求已取消"))
			return
		}
		if strings.Contains(err.Error(), "upstream") || strings.Contains(err.Error(), "OpenAI") || strings.Contains(err.Error(), "no available compatible OpenAI account") {
			response.ErrorFrom(c, infraerrors.ServiceUnavailable("PROMPT_OPTIMIZATION_UPSTREAM_FAILED", err.Error()))
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *ImageStudioHandler) DeleteTask(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	if err := h.service.DeleteTask(c.Request.Context(), subject.UserID, c.Param("task_id")); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(c, "task not found")
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *ImageStudioHandler) GetAssetContent(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	assetID, err := strconv.ParseInt(strings.TrimSpace(c.Param("asset_id")), 10, 64)
	if err != nil || assetID <= 0 {
		response.BadRequest(c, "invalid asset_id")
		return
	}
	reader, contentType, size, err := h.service.OpenAsset(c.Request.Context(), subject.UserID, assetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(c, "asset not found")
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	defer reader.Close()
	c.Header("Cache-Control", "private, max-age=3600")
	if size > 0 {
		c.Header("Content-Length", strconv.FormatInt(size, 10))
	}
	c.DataFromReader(http.StatusOK, size, contentType, reader, nil)
}
