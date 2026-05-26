package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"math"
	mathrand "math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

const (
	defaultImageStudioModel          = "gpt-image-2"
	defaultImageStudioPromptModel    = "gpt-5.5"
	defaultImageStudioQuality        = "high"
	defaultImageStudioPollInterval   = 3 * time.Second
	defaultImageStudioTaskTimeout    = 15 * time.Minute
	defaultImageStudioPromptTimeout  = 60 * time.Second
	defaultImageStudioPromptAttempt  = 45 * time.Second
	imageStudioPromptMaxOutputTokens = 420
	imageStudioLongPromptRunes       = 180
	defaultImageStudioWorkerCount    = 2
	imageStudioMaxPromptRunes        = 8000
	imageStudioMaxReferenceImageSize = openAIImageMaxUploadPartSize
	imageStudioMaxDownloadBytes      = 128 << 20
)

func ImageStudioMaxReferenceBytes() int64 {
	return imageStudioMaxReferenceImageSize
}

type ImageStudioService struct {
	repo                     ImageGenerationRepository
	apiKeyService            *APIKeyService
	gatewayService           *OpenAIGatewayService
	billingCacheService      *BillingCacheService
	subscriptionService      *SubscriptionService
	contentModerationService *ContentModerationService
	concurrencyService       *ConcurrencyService
	storage                  ImageStudioStorage
	cfg                      *config.Config

	workerCount  int
	pollInterval time.Duration
	taskTimeout  time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	wakeCh chan struct{}
	wg     sync.WaitGroup
}

type ImageStudioCreateTaskInput struct {
	UserID         int64
	APIKeyID       int64
	Model          string
	Prompt         string
	Ratio          string
	Resolution     string
	Quality        string
	Count          int
	ClientIP       string
	UserAgent      string
	ReferenceImage *ImageStudioUpload
}

type ImageStudioOptimizePromptInput struct {
	UserID         int64
	APIKeyID       int64
	Prompt         string
	Ratio          string
	Resolution     string
	Quality        string
	PreviousPrompt string
	Variant        int
	ClientIP       string
	UserAgent      string
}

type ImageStudioUpload struct {
	FileName    string
	ContentType string
	Data        []byte
}

type ImageStudioTaskOutput struct {
	ID            int64     `json:"id"`
	Seq           int       `json:"seq"`
	Kind          string    `json:"kind"`
	MimeType      string    `json:"mime_type"`
	Width         int       `json:"width"`
	Height        int       `json:"height"`
	SizeBytes     int64     `json:"size_bytes"`
	RevisedPrompt *string   `json:"revised_prompt,omitempty"`
	URL           string    `json:"url"`
	CreatedAt     time.Time `json:"created_at"`
}

type ImageStudioTaskResponse struct {
	ID           int64                   `json:"id"`
	TaskID       string                  `json:"task_id"`
	APIKeyID     int64                   `json:"api_key_id"`
	Mode         string                  `json:"mode"`
	Model        string                  `json:"model"`
	Prompt       string                  `json:"prompt"`
	Ratio        string                  `json:"ratio"`
	Resolution   string                  `json:"resolution"`
	Size         string                  `json:"size"`
	Quality      string                  `json:"quality"`
	Count        int                     `json:"count"`
	Status       string                  `json:"status"`
	Progress     int                     `json:"progress"`
	ErrorMessage *string                 `json:"error,omitempty"`
	StartedAt    *time.Time              `json:"started_at,omitempty"`
	FinishedAt   *time.Time              `json:"finished_at,omitempty"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
	Assets       []ImageStudioTaskOutput `json:"assets"`
}

type ImageStudioPromptOptimizationResponse struct {
	Prompt       string `json:"prompt"`
	SourcePrompt string `json:"source_prompt"`
	Model        string `json:"model"`
}

func NewImageStudioService(
	repo ImageGenerationRepository,
	apiKeyService *APIKeyService,
	gatewayService *OpenAIGatewayService,
	billingCacheService *BillingCacheService,
	subscriptionService *SubscriptionService,
	contentModerationService *ContentModerationService,
	concurrencyService *ConcurrencyService,
	storage ImageStudioStorage,
	cfg *config.Config,
) *ImageStudioService {
	ctx, cancel := context.WithCancel(context.Background())
	svc := &ImageStudioService{
		repo:                     repo,
		apiKeyService:            apiKeyService,
		gatewayService:           gatewayService,
		billingCacheService:      billingCacheService,
		subscriptionService:      subscriptionService,
		contentModerationService: contentModerationService,
		concurrencyService:       concurrencyService,
		storage:                  storage,
		cfg:                      cfg,
		workerCount:              defaultImageStudioWorkerCount,
		pollInterval:             defaultImageStudioPollInterval,
		taskTimeout:              defaultImageStudioTaskTimeout,
		ctx:                      ctx,
		cancel:                   cancel,
		wakeCh:                   make(chan struct{}, 1),
	}
	if cfg != nil {
		if cfg.ImageStudio.Worker.WorkerCount > 0 {
			svc.workerCount = cfg.ImageStudio.Worker.WorkerCount
		}
		if cfg.ImageStudio.Worker.PollIntervalSeconds > 0 {
			svc.pollInterval = time.Duration(cfg.ImageStudio.Worker.PollIntervalSeconds) * time.Second
		}
		if cfg.ImageStudio.Worker.TaskTimeoutSeconds > 0 {
			svc.taskTimeout = time.Duration(cfg.ImageStudio.Worker.TaskTimeoutSeconds) * time.Second
		}
	}
	return svc
}

func (s *ImageStudioService) Start() {
	if s == nil || s.repo == nil || s.workerCount <= 0 {
		return
	}
	if recovered, err := s.repo.RecoverRunningTasks(context.Background(), s.taskTimeout); err != nil {
		loggerLegacyImageStudio("recover running tasks failed: %v", err)
	} else if recovered > 0 {
		loggerLegacyImageStudio("recovered %d stale running image studio tasks", recovered)
	}
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.workerLoop(i + 1)
	}
	s.notify()
}

func (s *ImageStudioService) Stop() {
	if s == nil || s.cancel == nil {
		return
	}
	s.cancel()
	s.wg.Wait()
}

func (s *ImageStudioService) ListKeys(ctx context.Context, userID int64) ([]ImageStudioKeySummary, error) {
	keys, _, err := s.apiKeyService.List(ctx, userID, pagination.PaginationParams{Page: 1, PageSize: 1000}, APIKeyListFilters{Status: StatusActive})
	if err != nil {
		return nil, err
	}
	out := make([]ImageStudioKeySummary, 0, len(keys))
	for _, key := range keys {
		if key.UserID != userID || key.Status != StatusActive || key.Group == nil {
			continue
		}
		if key.Group.Platform != PlatformOpenAI || !GroupAllowsImageGeneration(key.Group) {
			continue
		}
		out = append(out, ImageStudioKeySummary{
			ID:        key.ID,
			Name:      key.Name,
			GroupID:   key.GroupID,
			GroupName: key.Group.Name,
			Platform:  key.Group.Platform,
			Status:    key.Status,
		})
	}
	return out, nil
}

func (s *ImageStudioService) CreateTask(ctx context.Context, input ImageStudioCreateTaskInput) (*ImageStudioTaskResponse, error) {
	apiKey, err := s.loadUsableImageStudioKey(ctx, input.UserID, input.APIKeyID)
	if err != nil {
		return nil, err
	}
	model := strings.TrimSpace(input.Model)
	if model == "" {
		model = defaultImageStudioModel
	}
	if err := validateOpenAIImagesModel(model); err != nil {
		return nil, err
	}
	prompt := strings.TrimSpace(input.Prompt)
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	if len([]rune(prompt)) > imageStudioMaxPromptRunes {
		return nil, fmt.Errorf("prompt is too long")
	}
	ratio, resolution, size := ResolveImageStudioSize(input.Ratio, input.Resolution)
	quality := strings.TrimSpace(input.Quality)
	if quality == "" {
		quality = defaultImageStudioQuality
	}
	count := input.Count
	if count <= 0 {
		count = 1
	}
	if count > 4 {
		count = 4
	}
	mode := ImageStudioModeTextToImage
	if input.ReferenceImage != nil && len(input.ReferenceImage.Data) > 0 {
		if len(input.ReferenceImage.Data) > imageStudioMaxReferenceImageSize {
			return nil, fmt.Errorf("reference image exceeds %d bytes", imageStudioMaxReferenceImageSize)
		}
		mode = ImageStudioModeImageToImage
	}
	requestMeta := mustMarshalJSON(map[string]any{
		"client_ip":  input.ClientIP,
		"user_agent": input.UserAgent,
	})
	task := &ImageGenerationTask{
		TaskID:      "img_" + strings.ReplaceAll(uuid.NewString(), "-", ""),
		UserID:      input.UserID,
		APIKeyID:    apiKey.ID,
		GroupID:     apiKey.GroupID,
		Mode:        mode,
		Model:       model,
		Prompt:      prompt,
		Ratio:       ratio,
		Resolution:  resolution,
		Size:        size,
		Quality:     quality,
		Count:       count,
		Status:      ImageStudioTaskStatusPending,
		Progress:    0,
		RequestMeta: requestMeta,
		UsageMeta:   mustMarshalJSON(map[string]any{"outputs": 0, "attempts": 0}),
	}
	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}
	if input.ReferenceImage != nil && len(input.ReferenceImage.Data) > 0 {
		asset, err := s.saveAsset(ctx, task, 0, ImageStudioAssetKindInput, input.ReferenceImage.FileName, input.ReferenceImage.ContentType, input.ReferenceImage.Data, nil, nil)
		if err != nil {
			_ = s.repo.SetTaskFailed(ctx, task.TaskID, "failed to cache reference image", nil)
			return nil, err
		}
		task.Assets = append(task.Assets, *asset)
	}
	s.notify()
	return imageStudioTaskToResponse(task), nil
}

func (s *ImageStudioService) GetTask(ctx context.Context, userID int64, taskID string) (*ImageStudioTaskResponse, error) {
	task, err := s.repo.GetTaskByTaskID(ctx, userID, strings.TrimSpace(taskID), false)
	if err != nil {
		return nil, err
	}
	return imageStudioTaskToResponse(task), nil
}

func (s *ImageStudioService) ListTasks(ctx context.Context, userID int64, params pagination.PaginationParams) ([]ImageStudioTaskResponse, *pagination.PaginationResult, error) {
	tasks, pag, err := s.repo.ListTasksByUserID(ctx, userID, params)
	if err != nil {
		return nil, nil, err
	}
	out := make([]ImageStudioTaskResponse, 0, len(tasks))
	for i := range tasks {
		out = append(out, *imageStudioTaskToResponse(&tasks[i]))
	}
	return out, pag, nil
}

func (s *ImageStudioService) OptimizePrompt(ctx context.Context, input ImageStudioOptimizePromptInput) (*ImageStudioPromptOptimizationResponse, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, defaultImageStudioPromptTimeout)
	defer cancel()
	apiKey, err := s.loadUsableImageStudioKey(ctx, input.UserID, input.APIKeyID)
	if err != nil {
		return nil, err
	}
	prompt := strings.TrimSpace(input.Prompt)
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	if len([]rune(prompt)) > imageStudioMaxPromptRunes {
		return nil, fmt.Errorf("prompt is too long")
	}
	var userRelease func()
	if s.concurrencyService != nil && apiKey.User != nil {
		acquired, err := s.concurrencyService.AcquireUserSlot(ctx, apiKey.User.ID, apiKey.User.Concurrency)
		if err != nil {
			return nil, err
		}
		if acquired == nil || !acquired.Acquired {
			return nil, fmt.Errorf("user concurrency limit reached")
		}
		userRelease = acquired.ReleaseFunc
	}
	if userRelease != nil {
		defer userRelease()
	}
	ratio, resolution, size := ResolveImageStudioSize(input.Ratio, input.Resolution)
	input.Ratio = ratio
	input.Resolution = resolution
	quality := strings.TrimSpace(input.Quality)
	if quality == "" {
		quality = defaultImageStudioQuality
	}
	input.Quality = quality
	model := defaultImageStudioPromptModel
	body, err := buildImageStudioPromptOptimizationRequestBody(input, model, size)
	if err != nil {
		return nil, err
	}
	var subscription *UserSubscription
	if apiKey.Group != nil && apiKey.Group.IsSubscriptionType() && s.subscriptionService != nil {
		subscription, err = s.subscriptionService.GetActiveSubscription(ctx, input.UserID, apiKey.Group.ID)
		if err != nil {
			return nil, err
		}
	}
	if s.billingCacheService != nil {
		if err := s.billingCacheService.CheckBillingEligibility(ctx, apiKey.User, apiKey, apiKey.Group, subscription, QuotaPlatform(ctx, apiKey)); err != nil {
			return nil, err
		}
	}
	if err := s.checkPromptOptimizationModeration(ctx, apiKey, model, body); err != nil {
		return nil, err
	}
	output, result, account, reqBody, err := s.executePromptOptimization(ctx, apiKey, body, model)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, newImageStudioPromptTimeoutError()
		}
		if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
			return nil, fmt.Errorf("prompt optimization request canceled")
		}
		return nil, err
	}
	output = cleanImageStudioOptimizedPrompt(output)
	if output == "" {
		return nil, fmt.Errorf("upstream response did not contain optimized prompt")
	}
	if result != nil && s.gatewayService != nil {
		channelMapping, _ := s.gatewayService.ResolveChannelMappingAndRestrict(ctx, apiKey.GroupID, model)
		if err := s.gatewayService.RecordUsage(ctx, &OpenAIRecordUsageInput{
			Result:             result,
			APIKey:             apiKey,
			User:               apiKey.User,
			Account:            account,
			Subscription:       subscription,
			InboundEndpoint:    "/api/v1/image-studio/prompt/optimize",
			UpstreamEndpoint:   openAIResponsesEndpoint,
			UserAgent:          input.UserAgent,
			IPAddress:          input.ClientIP,
			RequestPayloadHash: HashUsageRequestPayload(reqBody),
			APIKeyService:      s.apiKeyService,
			ChannelUsageFields: channelMapping.ToUsageFields(model, result.UpstreamModel),
		}); err != nil {
			loggerLegacyImageStudio("prompt optimization record usage failed: %v", err)
		}
	}
	return &ImageStudioPromptOptimizationResponse{
		Prompt:       output,
		SourcePrompt: prompt,
		Model:        model,
	}, nil
}

func (s *ImageStudioService) DeleteTask(ctx context.Context, userID int64, taskID string) error {
	return s.repo.SoftDeleteTask(ctx, userID, strings.TrimSpace(taskID))
}

func (s *ImageStudioService) OpenAsset(ctx context.Context, userID, assetID int64) (io.ReadCloser, string, int64, error) {
	asset, err := s.repo.GetAssetByIDForUser(ctx, userID, assetID)
	if err != nil {
		return nil, "", 0, err
	}
	return s.storage.Open(ctx, asset)
}

func (s *ImageStudioService) workerLoop(workerID int) {
	defer s.wg.Done()
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.wakeCh:
			s.drainPending(workerID)
		case <-ticker.C:
			s.drainPending(workerID)
		}
	}
}

func (s *ImageStudioService) drainPending(workerID int) {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}
		task, err := s.repo.ClaimNextPendingTask(s.ctx, s.taskTimeout)
		if errors.Is(err, ErrImageStudioNoPendingTask) {
			return
		}
		if err != nil {
			loggerLegacyImageStudio("worker %d claim task failed: %v", workerID, err)
			return
		}
		if task == nil {
			return
		}
		taskCtx, cancel := context.WithTimeout(s.ctx, s.taskTimeout)
		err = s.executeTask(taskCtx, task)
		cancel()
		if err != nil {
			loggerLegacyImageStudio("task %s failed: %v", task.TaskID, err)
			_ = s.repo.SetTaskFailed(context.Background(), task.TaskID, err.Error(), task.UsageMeta)
		}
	}
}

func (s *ImageStudioService) executeTask(ctx context.Context, task *ImageGenerationTask) error {
	apiKey, err := s.loadUsableImageStudioKey(ctx, task.UserID, task.APIKeyID)
	if err != nil {
		return err
	}
	var userRelease func()
	if s.concurrencyService != nil && apiKey.User != nil {
		acquired, err := s.concurrencyService.AcquireUserSlot(ctx, apiKey.User.ID, apiKey.User.Concurrency)
		if err != nil {
			return err
		}
		if acquired == nil || !acquired.Acquired {
			return fmt.Errorf("user concurrency limit reached")
		}
		userRelease = acquired.ReleaseFunc
	}
	if userRelease != nil {
		defer userRelease()
	}
	var subscription *UserSubscription
	if apiKey.Group != nil && apiKey.Group.IsSubscriptionType() && s.subscriptionService != nil {
		subscription, err = s.subscriptionService.GetActiveSubscription(ctx, task.UserID, apiKey.Group.ID)
		if err != nil {
			return err
		}
	}
	if s.billingCacheService != nil {
		if err := s.billingCacheService.CheckBillingEligibility(ctx, apiKey.User, apiKey, apiKey.Group, subscription, QuotaPlatform(ctx, apiKey)); err != nil {
			return err
		}
	}
	inputAssets, err := s.inputAssets(ctx, task.TaskID)
	if err != nil {
		return err
	}
	if task.Mode == ImageStudioModeImageToImage && len(inputAssets) == 0 {
		return fmt.Errorf("image edit task is missing reference image")
	}
	if err := s.checkModeration(ctx, task, apiKey, inputAssets); err != nil {
		return err
	}
	progressDone := s.startEstimatedProgress(ctx, task)
	defer progressDone()
	usageMeta := imageStudioUsageMeta{Outputs: 0, Attempts: 0, Results: []imageStudioUsageResult{}}
	for i := 0; i < task.Count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		progress := 10 + int(float64(i)/float64(task.Count)*80)
		_ = s.repo.UpdateTaskProgress(ctx, task.TaskID, progress, mustMarshalJSON(usageMeta))
		outputs, result, account, respBody, endpoint, reqBody, err := s.executeOneImage(ctx, task, apiKey, inputAssets)
		usageMeta.Attempts++
		if account != nil {
			accountID := account.ID
			task.AccountID = &accountID
		}
		usageMeta.Results = append(usageMeta.Results, imageStudioUsageResult{
			RequestID: resultRequestID(result),
			Model:     resultModel(result, task.Model),
			Size:      task.Size,
			ImageSize: resultImageSize(result, task.Resolution),
			Endpoint:  endpoint,
		})
		task.UsageMeta = mustMarshalJSON(usageMeta)
		_ = s.repo.AppendUpstreamLog(ctx, &ImageGenerationUpstreamLog{
			TaskID:          task.TaskID,
			UserID:          task.UserID,
			AccountID:       task.AccountID,
			Provider:        PlatformOpenAI,
			Endpoint:        endpoint,
			StatusCode:      recorderStatusFromBody(respBody, err),
			DurationMs:      durationMs(result),
			RequestExcerpt:  safePayloadExcerpt(reqBody),
			ResponseExcerpt: safePayloadExcerpt(respBody),
			ErrorMessage:    errStringPtr(err),
			Meta:            mustMarshalJSON(map[string]any{"iteration": i + 1}),
		})
		if err != nil {
			if len(outputs) == 0 {
				return err
			}
		}
		if len(outputs) == 0 {
			return fmt.Errorf("upstream response did not contain an image for item %d of %d", i+1, task.Count)
		}
		seqBase := usageMeta.Outputs
		for j, output := range outputs {
			asset, err := s.saveAsset(ctx, task, seqBase+j, ImageStudioAssetKindOutput, "", output.MimeType, output.Data, output.OriginalURL, output.RevisedPrompt)
			if err != nil {
				task.UsageMeta = mustMarshalJSON(usageMeta)
				return err
			}
			task.Assets = append(task.Assets, *asset)
			usageMeta.Outputs++
		}
		task.UsageMeta = mustMarshalJSON(usageMeta)
		iterationProgress := 10 + int(float64(usageMeta.Outputs)/float64(task.Count)*80)
		if iterationProgress > 90 {
			iterationProgress = 90
		}
		_ = s.repo.UpdateTaskProgress(ctx, task.TaskID, iterationProgress, task.UsageMeta)
		if result != nil && result.ImageCount > 0 && s.gatewayService != nil {
			channelMapping, _ := s.gatewayService.ResolveChannelMappingAndRestrict(ctx, apiKey.GroupID, task.Model)
			if err := s.gatewayService.RecordUsage(ctx, &OpenAIRecordUsageInput{
				Result:             result,
				APIKey:             apiKey,
				User:               apiKey.User,
				Account:            account,
				Subscription:       subscription,
				InboundEndpoint:    endpoint,
				UpstreamEndpoint:   endpoint,
				UserAgent:          taskMetaString(task.RequestMeta, "user_agent"),
				IPAddress:          taskMetaString(task.RequestMeta, "client_ip"),
				RequestPayloadHash: HashUsageRequestPayload(reqBody),
				APIKeyService:      s.apiKeyService,
				ChannelUsageFields: channelMapping.ToUsageFields(task.Model, result.UpstreamModel),
			}); err != nil {
				loggerLegacyImageStudio("task %s record usage failed: %v", task.TaskID, err)
			}
		}
	}
	if usageMeta.Outputs == 0 {
		return fmt.Errorf("upstream response did not contain images")
	}
	if usageMeta.Outputs < task.Count {
		return fmt.Errorf("generated %d of %d requested images", usageMeta.Outputs, task.Count)
	}
	task.UsageMeta = mustMarshalJSON(usageMeta)
	return s.repo.SetTaskSucceeded(ctx, task.TaskID, task.UsageMeta)
}

func (s *ImageStudioService) executeOneImage(ctx context.Context, task *ImageGenerationTask, apiKey *APIKey, inputAssets []ImageGenerationAsset) ([]imageStudioOutput, *OpenAIForwardResult, *Account, []byte, string, []byte, error) {
	body, contentType, endpoint, err := s.buildOpenAIImagesBody(ctx, task, inputAssets)
	if err != nil {
		return nil, nil, nil, nil, endpoint, nil, err
	}
	parsed, parseCtx, err := parseImageStudioForwardRequest(body, contentType, endpoint)
	if err != nil {
		return nil, nil, nil, nil, endpoint, body, err
	}
	channelMapping, _ := s.gatewayService.ResolveChannelMappingAndRestrict(ctx, apiKey.GroupID, parsed.Model)
	sessionHash := s.gatewayService.GenerateExplicitSessionHash(parseCtx, body)
	failed := map[int64]struct{}{}
	maxSwitches := 3
	if s.cfg != nil && s.cfg.Gateway.MaxAccountSwitches > 0 {
		maxSwitches = s.cfg.Gateway.MaxAccountSwitches
	}
	var lastErr error
	for attempt := 0; attempt <= maxSwitches; attempt++ {
		selection, _, err := s.gatewayService.SelectAccountWithSchedulerForImages(ctx, apiKey.GroupID, sessionHash, parsed.Model, failed, parsed.RequiredCapability)
		if err != nil {
			return nil, nil, nil, nil, endpoint, body, err
		}
		if selection == nil || selection.Account == nil {
			return nil, nil, nil, nil, endpoint, body, fmt.Errorf("no available compatible OpenAI account")
		}
		account := selection.Account
		release, acquired, err := s.acquireImageStudioAccountSlot(ctx, selection)
		if err != nil {
			return nil, nil, account, nil, endpoint, body, err
		}
		if !acquired {
			failed[account.ID] = struct{}{}
			lastErr = fmt.Errorf("account concurrency limit reached")
			continue
		}
		forwardCtx := cloneImageStudioGinContext(body, contentType, endpoint)
		forwardCtx.Set("api_key", apiKey)
		start := time.Now()
		result, err := s.gatewayService.ForwardImages(ctx, forwardCtx, account, body, parsed, channelMapping.MappedModel)
		duration := time.Since(start)
		if release != nil {
			release()
		}
		respBody := imageStudioRecorderBody(forwardCtx)
		if result != nil && result.Duration == 0 {
			result.Duration = duration
		}
		outputs, extractErr := extractImageStudioOutputs(ctx, respBody)
		if extractErr != nil && err == nil {
			err = extractErr
		}
		if err == nil || len(outputs) > 0 || extractErr != nil {
			s.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, true, resultFirstToken(result))
			return outputs, result, account, respBody, endpoint, body, err
		}
		var failoverErr *UpstreamFailoverError
		if errors.As(err, &failoverErr) {
			s.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
			failed[account.ID] = struct{}{}
			lastErr = err
			continue
		}
		return nil, result, account, respBody, endpoint, body, err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("image generation account failover exhausted")
	}
	return nil, nil, nil, nil, endpoint, body, lastErr
}

func (s *ImageStudioService) executePromptOptimization(ctx context.Context, apiKey *APIKey, body []byte, model string) (string, *OpenAIForwardResult, *Account, []byte, error) {
	if s == nil || s.gatewayService == nil {
		return "", nil, nil, body, fmt.Errorf("OpenAI gateway is not available")
	}
	channelMapping, _ := s.gatewayService.ResolveChannelMappingAndRestrict(ctx, apiKey.GroupID, model)
	sessionHash := imageStudioPromptOptimizationSessionHash(body)
	failed := map[int64]struct{}{}
	maxSwitches := 3
	if s.cfg != nil && s.cfg.Gateway.MaxAccountSwitches > 0 {
		maxSwitches = s.cfg.Gateway.MaxAccountSwitches
	}
	var lastErr error
	for attempt := 0; attempt <= maxSwitches; attempt++ {
		selection, err := s.selectPromptOptimizationAccount(ctx, apiKey.GroupID, sessionHash, model, failed)
		if err != nil {
			if lastErr != nil {
				if isImageStudioDeadlineError(lastErr) {
					return "", nil, nil, body, newImageStudioPromptTimeoutError()
				}
				return "", nil, nil, body, lastErr
			}
			return "", nil, nil, body, err
		}
		if selection == nil || selection.Account == nil {
			if lastErr != nil {
				if isImageStudioDeadlineError(lastErr) {
					return "", nil, nil, body, newImageStudioPromptTimeoutError()
				}
				return "", nil, nil, body, lastErr
			}
			return "", nil, nil, body, fmt.Errorf("no available compatible OpenAI account")
		}
		account := selection.Account
		release, acquired, err := s.acquireImageStudioAccountSlot(ctx, selection)
		if err != nil {
			return "", nil, account, body, err
		}
		if !acquired {
			failed[account.ID] = struct{}{}
			lastErr = fmt.Errorf("account concurrency limit reached")
			continue
		}
		forwardBody := body
		if channelMapping.Mapped {
			forwardBody = s.gatewayService.ReplaceModelInBody(body, channelMapping.MappedModel)
		}
		forwardCtx := cloneImageStudioGinContext(forwardBody, "application/json", openAIResponsesEndpoint)
		forwardCtx.Set("api_key", apiKey)
		prepareImageStudioPromptOptimizationContext(forwardCtx, account)
		SetOpenAIClientTransport(forwardCtx, OpenAIClientTransportHTTP)
		start := time.Now()
		attemptCtx, attemptCancel := context.WithTimeout(ctx, defaultImageStudioPromptAttempt)
		forwardCallCtx := withAttachedUpstreamContext(attemptCtx)
		loggerLegacyImageStudio(
			"prompt optimization attempt started: attempt=%d account_id=%d account_type=%s platform=%s model=%s mapped_model=%s",
			attempt+1,
			account.ID,
			account.Type,
			account.Platform,
			model,
			resultPromptOptimizationMappedModel(channelMapping, model),
		)
		result, err := s.gatewayService.Forward(forwardCallCtx, forwardCtx, account, forwardBody)
		attemptCancel()
		duration := time.Since(start)
		if release != nil {
			release()
		}
		if result != nil && result.Duration == 0 {
			result.Duration = duration
		}
		respBody := imageStudioRecorderBody(forwardCtx)
		output := extractImageStudioPromptText(respBody)
		if err == nil && strings.TrimSpace(output) != "" {
			s.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, true, resultFirstToken(result))
			loggerLegacyImageStudio(
				"prompt optimization attempt succeeded: attempt=%d account_id=%d duration_ms=%d model=%s mapped_model=%s",
				attempt+1,
				account.ID,
				duration.Milliseconds(),
				model,
				resultPromptOptimizationMappedModel(channelMapping, model),
			)
			return output, result, account, forwardBody, nil
		}
		if err == nil {
			err = fmt.Errorf("upstream response did not contain optimized prompt")
		}
		loggerLegacyImageStudio(
			"prompt optimization attempt failed: attempt=%d account_id=%d duration_ms=%d model=%s mapped_model=%s err=%v",
			attempt+1,
			account.ID,
			duration.Milliseconds(),
			model,
			resultPromptOptimizationMappedModel(channelMapping, model),
			err,
		)
		if isImageStudioDeadlineError(err) || errors.Is(attemptCtx.Err(), context.DeadlineExceeded) {
			s.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
			failed[account.ID] = struct{}{}
			lastErr = err
			if ctx.Err() != nil {
				return output, result, account, forwardBody, err
			}
			continue
		}
		var failoverErr *UpstreamFailoverError
		if errors.As(err, &failoverErr) {
			s.gatewayService.ReportOpenAIAccountScheduleResult(account.ID, false, nil)
			failed[account.ID] = struct{}{}
			lastErr = err
			continue
		}
		return output, result, account, forwardBody, err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("prompt optimization account failover exhausted")
	}
	if isImageStudioDeadlineError(lastErr) {
		return "", nil, nil, body, newImageStudioPromptTimeoutError()
	}
	return "", nil, nil, body, lastErr
}

func resultPromptOptimizationMappedModel(channelMapping ChannelMappingResult, fallback string) string {
	if channelMapping.MappedModel != "" {
		return channelMapping.MappedModel
	}
	return fallback
}

func (s *ImageStudioService) selectPromptOptimizationAccount(ctx context.Context, groupID *int64, sessionHash string, model string, failed map[int64]struct{}) (*AccountSelectionResult, error) {
	selection, _, err := s.gatewayService.SelectAccountWithScheduler(
		ctx,
		groupID,
		"",
		sessionHash,
		model,
		failed,
		OpenAIUpstreamTransportHTTPSSE,
		false,
	)
	if err == nil || len(failed) > 0 {
		return selection, err
	}
	selection, _, fallbackErr := s.gatewayService.SelectAccountWithScheduler(
		ctx,
		groupID,
		"",
		sessionHash,
		"",
		failed,
		OpenAIUpstreamTransportHTTPSSE,
		false,
	)
	if fallbackErr != nil {
		return nil, err
	}
	return selection, nil
}

func (s *ImageStudioService) startEstimatedProgress(ctx context.Context, task *ImageGenerationTask) func() {
	if s == nil || s.repo == nil || task == nil || task.TaskID == "" {
		return func() {}
	}
	progressCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.runEstimatedProgress(progressCtx, task)
	}()
	return func() {
		cancel()
		<-done
	}
}

func (s *ImageStudioService) runEstimatedProgress(ctx context.Context, task *ImageGenerationTask) {
	estimate := imageStudioEstimatedDuration(task)
	if estimate <= 0 {
		estimate = 40 * time.Second
	}
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	start := time.Now()
	lastProgress := 10
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			progress := imageStudioEstimatedProgress(task, now.Sub(start), estimate)
			if progress <= lastProgress {
				continue
			}
			lastProgress = progress
			if err := s.repo.UpdateTaskProgressOnly(ctx, task.TaskID, progress); err != nil {
				loggerLegacyImageStudio("task %s progress update failed: %v", task.TaskID, err)
			}
			if progress >= 90 {
				return
			}
		}
	}
}

func imageStudioEstimatedProgress(task *ImageGenerationTask, elapsed time.Duration, estimate time.Duration) int {
	if elapsed <= 0 || estimate <= 0 {
		return 10
	}
	ratio := float64(elapsed) / float64(estimate)
	if ratio >= 1 {
		return 90
	}
	if ratio < 0 {
		ratio = 0
	}
	curved := 1 - math.Pow(1-ratio, 1.45)
	progress := 10 + int(math.Round(curved*80))
	progress += imageStudioProgressJitter(task, elapsed)
	if progress < 10 {
		return 10
	}
	if progress > 90 {
		return 90
	}
	return progress
}

func imageStudioProgressJitter(task *ImageGenerationTask, elapsed time.Duration) int {
	if task == nil {
		return 0
	}
	slot := int(elapsed / (2 * time.Second))
	if slot <= 0 {
		return 0
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(fmt.Sprintf("%s:%d", task.TaskID, slot)))
	rng := mathrand.New(mathrand.NewSource(int64(h.Sum64())))
	return rng.Intn(5) - 1
}

func imageStudioEstimatedDuration(task *ImageGenerationTask) time.Duration {
	if task == nil {
		return 40 * time.Second
	}
	count := task.Count
	if count <= 0 {
		count = 1
	}
	var perImage time.Duration
	resolution := strings.ToUpper(strings.TrimSpace(task.Resolution))
	quality := strings.ToLower(strings.TrimSpace(task.Quality))
	switch resolution {
	case "4K":
		perImage = 150 * time.Second
		if quality == "high" {
			perImage = 165 * time.Second
		}
		if quality == "low" {
			perImage = 120 * time.Second
		}
	case "2K":
		perImage = 80 * time.Second
		if quality == "high" {
			perImage = 88 * time.Second
		}
		if quality == "low" {
			perImage = 64 * time.Second
		}
	default:
		perImage = 40 * time.Second
		if quality == "low" {
			perImage = 32 * time.Second
		}
	}
	return time.Duration(count) * perImage
}

func (s *ImageStudioService) buildOpenAIImagesBody(ctx context.Context, task *ImageGenerationTask, inputAssets []ImageGenerationAsset) ([]byte, string, string, error) {
	if task.Mode != ImageStudioModeImageToImage {
		payload := map[string]any{
			"model":           task.Model,
			"prompt":          task.Prompt,
			"size":            task.Size,
			"n":               1,
			"quality":         task.Quality,
			"response_format": "b64_json",
		}
		body, err := json.Marshal(payload)
		return body, "application/json", openAIImagesGenerationsEndpoint, err
	}
	if len(inputAssets) == 0 {
		return nil, "", openAIImagesEditsEndpoint, fmt.Errorf("reference image is required")
	}
	reader, mimeType, _, err := s.storage.Open(ctx, &inputAssets[0])
	if err != nil {
		return nil, "", openAIImagesEditsEndpoint, err
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, imageStudioMaxReferenceImageSize+1))
	if err != nil {
		return nil, "", openAIImagesEditsEndpoint, err
	}
	if len(data) > imageStudioMaxReferenceImageSize {
		return nil, "", openAIImagesEditsEndpoint, fmt.Errorf("reference image exceeds %d bytes", imageStudioMaxReferenceImageSize)
	}
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	fields := map[string]string{
		"model":           task.Model,
		"prompt":          task.Prompt,
		"size":            task.Size,
		"n":               "1",
		"quality":         task.Quality,
		"response_format": "b64_json",
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", openAIImagesEditsEndpoint, err
		}
	}
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="image"; filename="reference.png"`)
	header.Set("Content-Type", mimeType)
	part, err := writer.CreatePart(header)
	if err != nil {
		return nil, "", openAIImagesEditsEndpoint, err
	}
	if _, err := part.Write(data); err != nil {
		return nil, "", openAIImagesEditsEndpoint, err
	}
	if err := writer.Close(); err != nil {
		return nil, "", openAIImagesEditsEndpoint, err
	}
	return buf.Bytes(), writer.FormDataContentType(), openAIImagesEditsEndpoint, nil
}

func (s *ImageStudioService) checkModeration(ctx context.Context, task *ImageGenerationTask, apiKey *APIKey, inputAssets []ImageGenerationAsset) error {
	if s.contentModerationService == nil {
		return nil
	}
	body, contentType, endpoint, err := s.buildOpenAIImagesBody(ctx, task, inputAssets)
	if err != nil {
		return err
	}
	parsed, _, err := parseImageStudioForwardRequest(body, contentType, endpoint)
	if err != nil {
		return err
	}
	groupName := ""
	if apiKey.Group != nil {
		groupName = apiKey.Group.Name
	}
	userEmail := ""
	if apiKey.User != nil {
		userEmail = apiKey.User.Email
	}
	decision, err := s.contentModerationService.Check(ctx, ContentModerationCheckInput{
		RequestID:  task.TaskID,
		UserID:     task.UserID,
		UserEmail:  userEmail,
		APIKeyID:   apiKey.ID,
		APIKeyName: apiKey.Name,
		GroupID:    apiKey.GroupID,
		GroupName:  groupName,
		Endpoint:   endpoint,
		Provider:   PlatformOpenAI,
		Model:      parsed.Model,
		Protocol:   ContentModerationProtocolOpenAIImages,
		Body:       parsed.ModerationBody(),
	})
	if err != nil {
		return err
	}
	if decision != nil && decision.Blocked {
		return fmt.Errorf("%s", decision.Message)
	}
	return nil
}

func (s *ImageStudioService) checkPromptOptimizationModeration(ctx context.Context, apiKey *APIKey, model string, body []byte) error {
	if s.contentModerationService == nil || apiKey == nil {
		return nil
	}
	groupName := ""
	if apiKey.Group != nil {
		groupName = apiKey.Group.Name
	}
	userEmail := ""
	if apiKey.User != nil {
		userEmail = apiKey.User.Email
	}
	decision, err := s.contentModerationService.Check(ctx, ContentModerationCheckInput{
		RequestID:  "image_studio_prompt_optimize",
		UserID:     apiKey.UserID,
		UserEmail:  userEmail,
		APIKeyID:   apiKey.ID,
		APIKeyName: apiKey.Name,
		GroupID:    apiKey.GroupID,
		GroupName:  groupName,
		Endpoint:   openAIResponsesEndpoint,
		Provider:   PlatformOpenAI,
		Model:      model,
		Protocol:   ContentModerationProtocolOpenAIResponses,
		Body:       body,
	})
	if err != nil {
		return err
	}
	if decision != nil && decision.Blocked {
		return fmt.Errorf("%s", decision.Message)
	}
	return nil
}

func (s *ImageStudioService) acquireImageStudioAccountSlot(ctx context.Context, selection *AccountSelectionResult) (func(), bool, error) {
	if selection == nil {
		return nil, false, nil
	}
	if selection.Acquired {
		return selection.ReleaseFunc, true, nil
	}
	if selection.WaitPlan == nil || s.concurrencyService == nil {
		return nil, false, nil
	}
	result, err := s.concurrencyService.AcquireAccountSlot(ctx, selection.WaitPlan.AccountID, selection.WaitPlan.MaxConcurrency)
	if err != nil {
		return nil, false, err
	}
	if result == nil || !result.Acquired {
		return nil, false, nil
	}
	return result.ReleaseFunc, true, nil
}

func (s *ImageStudioService) inputAssets(ctx context.Context, taskID string) ([]ImageGenerationAsset, error) {
	assets, err := s.repo.ListAssetsByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	out := make([]ImageGenerationAsset, 0, len(assets))
	for _, asset := range assets {
		if asset.Kind == ImageStudioAssetKindInput {
			out = append(out, asset)
		}
	}
	return out, nil
}

func (s *ImageStudioService) saveAsset(ctx context.Context, task *ImageGenerationTask, seq int, kind string, fileName string, mimeType string, data []byte, originalURL *string, revisedPrompt *string) (*ImageGenerationAsset, error) {
	stored, err := s.storage.Save(ctx, ImageStudioStoreInput{
		TaskID:        task.TaskID,
		UserID:        task.UserID,
		Seq:           seq,
		Kind:          kind,
		FileName:      fileName,
		MimeType:      mimeType,
		Data:          data,
		OriginalURL:   originalURL,
		RevisedPrompt: revisedPrompt,
		Meta:          mustMarshalJSON(map[string]any{"size": task.Size, "ratio": task.Ratio, "resolution": task.Resolution}),
	})
	if err != nil {
		return nil, err
	}
	asset := &ImageGenerationAsset{
		TaskID:        task.TaskID,
		UserID:        task.UserID,
		Seq:           seq,
		Kind:          kind,
		StorageDriver: stored.StorageDriver,
		StorageKey:    stored.StorageKey,
		MimeType:      stored.MimeType,
		Width:         stored.Width,
		Height:        stored.Height,
		SizeBytes:     stored.SizeBytes,
		OriginalURL:   originalURL,
		RevisedPrompt: revisedPrompt,
		Meta:          mustMarshalJSON(map[string]any{"size": task.Size, "ratio": task.Ratio, "resolution": task.Resolution}),
	}
	if err := s.repo.AddAsset(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}

func (s *ImageStudioService) loadUsableImageStudioKey(ctx context.Context, userID int64, apiKeyID int64) (*APIKey, error) {
	if apiKeyID <= 0 {
		return nil, fmt.Errorf("api_key_id is required")
	}
	apiKey, err := s.apiKeyService.GetByID(ctx, apiKeyID)
	if err != nil {
		return nil, err
	}
	if apiKey.UserID != userID {
		return nil, fmt.Errorf("api key does not belong to current user")
	}
	if apiKey.Status != StatusActive || apiKey.IsExpired() || apiKey.IsQuotaExhausted() {
		return nil, fmt.Errorf("api key is not active")
	}
	if apiKey.Group == nil || apiKey.Group.Platform != PlatformOpenAI {
		return nil, fmt.Errorf("api key must be bound to an OpenAI group")
	}
	if !GroupAllowsImageGeneration(apiKey.Group) {
		return nil, errors.New(ImageGenerationPermissionMessage())
	}
	if apiKey.User == nil {
		apiKey.User = &User{ID: userID}
	}
	_ = s.apiKeyService.TouchLastUsed(ctx, apiKey.ID)
	return apiKey, nil
}

func (s *ImageStudioService) notify() {
	select {
	case s.wakeCh <- struct{}{}:
	default:
	}
}

type imageStudioUsageMeta struct {
	Outputs  int                      `json:"outputs"`
	Attempts int                      `json:"attempts"`
	Results  []imageStudioUsageResult `json:"results"`
}

type imageStudioUsageResult struct {
	RequestID string `json:"request_id,omitempty"`
	Model     string `json:"model,omitempty"`
	Size      string `json:"size,omitempty"`
	ImageSize string `json:"image_size,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
}

type imageStudioOutput struct {
	Data          []byte
	MimeType      string
	OriginalURL   *string
	RevisedPrompt *string
}

func parseImageStudioForwardRequest(body []byte, contentType string, endpoint string) (*OpenAIImagesRequest, *gin.Context, error) {
	c := cloneImageStudioGinContext(body, contentType, endpoint)
	svc := &OpenAIGatewayService{}
	parsed, err := svc.ParseOpenAIImagesRequest(c, body)
	if err != nil {
		return nil, c, err
	}
	return parsed, c, nil
}

func cloneImageStudioGinContext(body []byte, contentType string, endpoint string) *gin.Context {
	rec := newGinResponseRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", "sub2api-image-studio")
	c.Request = req
	c.Set("image_studio_recorder", rec)
	return c
}

func prepareImageStudioPromptOptimizationContext(c *gin.Context, account *Account) {
	if c == nil || account == nil || account.Type != AccountTypeOAuth {
		return
	}
	c.Request.Header.Set("User-Agent", codexCLIUserAgent)
	c.Request.Header.Set("originator", "codex_cli_rs")
	c.Request.Header.Set("version", codexCLIVersion)
}

func imageStudioRecorderBody(c *gin.Context) []byte {
	if c == nil {
		return nil
	}
	if value, ok := c.Get("image_studio_recorder"); ok {
		if rec, ok := value.(*ginResponseRecorder); ok {
			return rec.BodyBytes()
		}
	}
	return nil
}

func extractImageStudioOutputs(ctx context.Context, body []byte) ([]imageStudioOutput, error) {
	if outputs, err := extractImageStudioOpenAIDataOutputs(ctx, body); err != nil || len(outputs) > 0 {
		return outputs, err
	}
	pointers := collectOpenAIImageInlineAssets(body, "")
	outputs := make([]imageStudioOutput, 0, len(pointers))
	errs := []error{}
	for _, pointer := range pointers {
		data, mimeType, originalURL, err := imageStudioPointerBytes(ctx, pointer)
		if err != nil || len(data) == 0 {
			if err == nil {
				err = fmt.Errorf("inline image asset is empty")
			}
			errs = append(errs, err)
			continue
		}
		prompt := strings.TrimSpace(pointer.Prompt)
		var revised *string
		if prompt != "" {
			revised = &prompt
		}
		outputs = append(outputs, imageStudioOutput{
			Data:          data,
			MimeType:      mimeType,
			OriginalURL:   originalURL,
			RevisedPrompt: revised,
		})
	}
	if len(outputs) == 0 && len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return outputs, nil
}

func extractImageStudioOpenAIDataOutputs(ctx context.Context, body []byte) ([]imageStudioOutput, error) {
	if len(body) == 0 {
		return nil, nil
	}
	var response struct {
		Data []struct {
			B64JSON       string `json:"b64_json"`
			URL           string `json:"url"`
			RevisedPrompt string `json:"revised_prompt"`
			MimeType      string `json:"mime_type"`
			ContentType   string `json:"content_type"`
			OutputFormat  string `json:"output_format"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, nil
	}
	if len(response.Data) == 0 {
		return nil, nil
	}
	outputs := make([]imageStudioOutput, 0, len(response.Data))
	errs := []error{}
	for _, item := range response.Data {
		mimeType := firstNonEmptyImageStudioString(item.MimeType, item.ContentType, openAIImageOutputMIMEType(item.OutputFormat))
		data, resolvedMimeType, originalURL, err := imageStudioPointerBytes(ctx, openAIImagePointerInfo{
			DownloadURL: strings.TrimSpace(item.URL),
			B64JSON:     strings.TrimSpace(item.B64JSON),
			MimeType:    mimeType,
			Prompt:      strings.TrimSpace(item.RevisedPrompt),
		})
		if err != nil || len(data) == 0 {
			if err == nil {
				err = fmt.Errorf("OpenAI image data is empty")
			}
			errs = append(errs, err)
			continue
		}
		var revised *string
		if prompt := strings.TrimSpace(item.RevisedPrompt); prompt != "" {
			revised = &prompt
		}
		outputs = append(outputs, imageStudioOutput{
			Data:          data,
			MimeType:      resolvedMimeType,
			OriginalURL:   originalURL,
			RevisedPrompt: revised,
		})
	}
	if len(outputs) == 0 && len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return outputs, nil
}

func imageStudioPointerBytes(ctx context.Context, pointer openAIImagePointerInfo) ([]byte, string, *string, error) {
	if normalized := normalizeOpenAIImageBase64(pointer.B64JSON); normalized != "" {
		data, err := base64.StdEncoding.DecodeString(normalized)
		return data, normalizeImageStudioMimeType(pointer.MimeType, "", data), nil, err
	}
	url := strings.TrimSpace(pointer.DownloadURL)
	if strings.HasPrefix(strings.ToLower(url), "data:image/") {
		idx := strings.Index(url, ",")
		if idx < 0 || idx+1 >= len(url) {
			return nil, "", nil, fmt.Errorf("invalid data url")
		}
		header := url[:idx]
		data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(url[idx+1:]))
		if err != nil {
			return nil, "", nil, err
		}
		mimeType := pointer.MimeType
		if semi := strings.Index(header, ";"); semi > 5 {
			mimeType = header[5:semi]
		}
		return data, normalizeImageStudioMimeType(mimeType, "", data), nil, nil
	}
	if strings.HasPrefix(strings.ToLower(url), "http://") || strings.HasPrefix(strings.ToLower(url), "https://") {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, "", nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, "", nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return nil, "", nil, fmt.Errorf("download image failed: %d", resp.StatusCode)
		}
		data, err := io.ReadAll(io.LimitReader(resp.Body, imageStudioMaxDownloadBytes+1))
		if err != nil {
			return nil, "", nil, err
		}
		if len(data) > imageStudioMaxDownloadBytes {
			return nil, "", nil, fmt.Errorf("downloaded image exceeds %d bytes", imageStudioMaxDownloadBytes)
		}
		originalURL := url
		return data, normalizeImageStudioMimeType(firstNonEmptyImageStudioString(pointer.MimeType, resp.Header.Get("Content-Type")), "", data), &originalURL, nil
	}
	return nil, "", nil, fmt.Errorf("unsupported image pointer")
}

func buildImageStudioPromptOptimizationRequestBody(input ImageStudioOptimizePromptInput, model string, size string) ([]byte, error) {
	return json.Marshal(map[string]any{
		"model":        model,
		"instructions": "You are a fast image prompt rewriting assistant. Return only the optimized image prompt text.",
		"input":        buildImageStudioPromptOptimizationMessages(input, size),
		"reasoning": map[string]string{
			"effort": "low",
		},
		"text": map[string]string{
			"verbosity": "low",
		},
		"max_output_tokens": imageStudioPromptMaxOutputTokens,
		"stream":            false,
	})
}

func buildImageStudioPromptOptimizationInput(input ImageStudioOptimizePromptInput, size string) string {
	var b strings.Builder
	variant := input.Variant
	if variant <= 0 {
		variant = 1
	}
	prompt := strings.TrimSpace(input.Prompt)
	b.WriteString("你是专业 AI 图像提示词优化助手。请基于用户原始提示词，快速输出一个更适合图像生成模型的提示词。\n")
	b.WriteString("要求：只输出优化后的提示词正文，不要 Markdown，不要解释，不要编号；保留用户核心意图；语言尽量跟随用户原始输入。\n")
	if len([]rune(prompt)) >= imageStudioLongPromptRunes {
		b.WriteString("原始提示词已经比较详细，请以整理、去重、压缩和强化画面一致性为主，不要继续堆叠细节；输出控制在 120-220 个中文字符左右。\n")
	} else {
		b.WriteString("原始提示词较短，请适度补充画面主体、环境、构图、光线、材质、风格和细节；输出控制在 120-220 个中文字符左右。\n")
	}
	b.WriteString(fmt.Sprintf("当前生成参数：比例 %s，分辨率 %s，实际尺寸 %s，质量 %s，候选版本 %d。\n", input.Ratio, input.Resolution, size, input.Quality, variant))
	if previous := strings.TrimSpace(input.PreviousPrompt); previous != "" {
		b.WriteString("上一版候选提示词如下。本次请生成明显不同但同样高质量的表达，避免重复相同句式和细节。\n")
		b.WriteString(previous)
		b.WriteString("\n")
	}
	b.WriteString("用户原始提示词：\n")
	b.WriteString(prompt)
	return b.String()
}

func buildImageStudioPromptOptimizationMessages(input ImageStudioOptimizePromptInput, size string) []map[string]any {
	return []map[string]any{
		{
			"type": "message",
			"role": "user",
			"content": []map[string]string{
				{
					"type": "input_text",
					"text": buildImageStudioPromptOptimizationInput(input, size),
				},
			},
		},
	}
}

func imageStudioPromptOptimizationSessionHash(body []byte) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte("image-studio-prompt:"))
	_, _ = h.Write(body)
	return fmt.Sprintf("image-studio-prompt-%x", h.Sum64())
}

func extractImageStudioPromptText(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	if text := extractImageStudioPromptTextFromJSON(body); text != "" {
		return text
	}
	return extractImageStudioPromptTextFromSSE(string(body))
}

func extractImageStudioPromptTextFromJSON(body []byte) string {
	if !gjson.ValidBytes(body) {
		return ""
	}
	for _, path := range []string{
		"output_text",
		"response.output_text",
		"choices.0.message.content",
		"message.content",
	} {
		value := gjson.GetBytes(body, path)
		if value.Type != gjson.String {
			continue
		}
		if text := strings.TrimSpace(value.String()); text != "" {
			return text
		}
	}
	var parts []string
	collectImageStudioPromptTextParts(gjson.GetBytes(body, "output"), &parts)
	collectImageStudioPromptTextParts(gjson.GetBytes(body, "response.output"), &parts)
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func collectImageStudioPromptTextParts(value gjson.Result, parts *[]string) {
	switch {
	case !value.Exists():
		return
	case value.Type == gjson.String:
		if text := strings.TrimSpace(value.String()); text != "" {
			*parts = append(*parts, text)
		}
	case value.IsArray():
		value.ForEach(func(_, item gjson.Result) bool {
			collectImageStudioPromptTextParts(item, parts)
			return true
		})
	case value.IsObject():
		typ := strings.ToLower(strings.TrimSpace(value.Get("type").String()))
		if typ == "output_text" || typ == "text" || typ == "input_text" || typ == "" {
			textValue := value.Get("text")
			if textValue.Type == gjson.String {
				text := strings.TrimSpace(textValue.String())
				if text == "" {
					return
				}
				*parts = append(*parts, text)
			}
		}
		collectImageStudioPromptTextParts(value.Get("content"), parts)
	}
}

func extractImageStudioPromptTextFromSSE(bodyText string) string {
	var deltas []string
	scanner := bufio.NewScanner(strings.NewReader(bodyText))
	scanner.Buffer(make([]byte, 0, 64*1024), defaultMaxLineSize)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		data, ok := extractOpenAISSEDataLine(line)
		if !ok {
			continue
		}
		data = strings.TrimSpace(data)
		if data == "" || data == "[DONE]" {
			continue
		}
		payload := []byte(data)
		eventType := strings.TrimSpace(gjson.GetBytes(payload, "type").String())
		if eventType == "response.output_text.delta" {
			if delta := gjson.GetBytes(payload, "delta").String(); delta != "" {
				deltas = append(deltas, delta)
			}
			continue
		}
		if eventType == "response.completed" {
			if text := extractImageStudioPromptTextFromJSON([]byte(gjson.GetBytes(payload, "response").Raw)); text != "" {
				return text
			}
		}
	}
	if len(deltas) > 0 {
		return strings.TrimSpace(strings.Join(deltas, ""))
	}
	return ""
}

func cleanImageStudioOptimizedPrompt(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "```text")
	value = strings.TrimPrefix(value, "```markdown")
	value = strings.TrimPrefix(value, "```")
	value = strings.TrimSuffix(value, "```")
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "\"'“”‘’")
	return strings.TrimSpace(value)
}

func newImageStudioPromptTimeoutError() error {
	return fmt.Errorf("prompt optimization timed out after %s", defaultImageStudioPromptTimeout)
}

func isImageStudioDeadlineError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "context deadline exceeded")
}

func imageStudioTaskToResponse(task *ImageGenerationTask) *ImageStudioTaskResponse {
	if task == nil {
		return nil
	}
	assets := make([]ImageStudioTaskOutput, 0, len(task.Assets))
	for _, asset := range task.Assets {
		if asset.Kind != ImageStudioAssetKindOutput {
			continue
		}
		assets = append(assets, ImageStudioTaskOutput{
			ID:            asset.ID,
			Seq:           asset.Seq,
			Kind:          asset.Kind,
			MimeType:      asset.MimeType,
			Width:         asset.Width,
			Height:        asset.Height,
			SizeBytes:     asset.SizeBytes,
			RevisedPrompt: asset.RevisedPrompt,
			URL:           fmt.Sprintf("/api/v1/image-studio/assets/%d/content", asset.ID),
			CreatedAt:     asset.CreatedAt,
		})
	}
	return &ImageStudioTaskResponse{
		ID:           task.ID,
		TaskID:       task.TaskID,
		APIKeyID:     task.APIKeyID,
		Mode:         task.Mode,
		Model:        task.Model,
		Prompt:       task.Prompt,
		Ratio:        task.Ratio,
		Resolution:   task.Resolution,
		Size:         task.Size,
		Quality:      task.Quality,
		Count:        task.Count,
		Status:       task.Status,
		Progress:     task.Progress,
		ErrorMessage: task.ErrorMessage,
		StartedAt:    task.StartedAt,
		FinishedAt:   task.FinishedAt,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
		Assets:       assets,
	}
}

func mustMarshalJSON(value any) json.RawMessage {
	body, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return body
}

func taskMetaString(raw json.RawMessage, key string) string {
	if len(raw) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	if value, ok := payload[key].(string); ok {
		return value
	}
	return ""
}

func safePayloadExcerpt(body []byte) string {
	const max = 4000
	if len(body) == 0 {
		return ""
	}
	if len(body) <= max {
		return string(body)
	}
	return string(body[:max]) + "..."
}

func errStringPtr(err error) *string {
	if err == nil {
		return nil
	}
	s := err.Error()
	return &s
}

func resultRequestID(result *OpenAIForwardResult) string {
	if result == nil {
		return ""
	}
	return result.RequestID
}

func resultModel(result *OpenAIForwardResult, fallback string) string {
	if result == nil || result.Model == "" {
		return fallback
	}
	return result.Model
}

func resultImageSize(result *OpenAIForwardResult, fallback string) string {
	if result == nil || result.ImageSize == "" {
		return fallback
	}
	return result.ImageSize
}

func durationMs(result *OpenAIForwardResult) int64 {
	if result == nil {
		return 0
	}
	return result.Duration.Milliseconds()
}

func resultFirstToken(result *OpenAIForwardResult) *int {
	if result == nil {
		return nil
	}
	return result.FirstTokenMs
}

func recorderStatusFromBody(body []byte, err error) *int {
	status := 200
	if err != nil {
		status = 502
	}
	var payload struct {
		Error any `json:"error"`
	}
	if len(body) > 0 && json.Unmarshal(body, &payload) == nil && payload.Error != nil {
		status = 502
	}
	return &status
}

func firstNonEmptyImageStudioString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func loggerLegacyImageStudio(format string, args ...any) {
	logger.LegacyPrintf("service.image_studio", format, args...)
}

type ginResponseRecorder struct {
	*httptest.ResponseRecorder
	status int
	size   int
}

func newGinResponseRecorder() *ginResponseRecorder {
	return &ginResponseRecorder{ResponseRecorder: httptest.NewRecorder(), status: http.StatusOK, size: -1}
}

func (r *ginResponseRecorder) CloseNotify() <-chan bool {
	ch := make(chan bool, 1)
	return ch
}

func (r *ginResponseRecorder) Flush() {}

func (r *ginResponseRecorder) Status() int {
	return r.status
}

func (r *ginResponseRecorder) Size() int {
	return r.size
}

func (r *ginResponseRecorder) Written() bool {
	return r.size >= 0
}

func (r *ginResponseRecorder) WriteHeaderNow() {
	if !r.Written() {
		r.size = 0
		r.ResponseRecorder.WriteHeader(r.status)
	}
}

func (r *ginResponseRecorder) WriteHeader(code int) {
	if code > 0 && !r.Written() {
		r.status = code
	}
}

func (r *ginResponseRecorder) Write(data []byte) (int, error) {
	r.WriteHeaderNow()
	n, err := r.ResponseRecorder.Write(data)
	r.size += n
	return n, err
}

func (r *ginResponseRecorder) WriteString(s string) (int, error) {
	r.WriteHeaderNow()
	n, err := r.ResponseRecorder.WriteString(s)
	r.size += n
	return n, err
}

func (r *ginResponseRecorder) BodyBytes() []byte {
	if r.Body == nil {
		return nil
	}
	return r.Body.Bytes()
}

func (r *ginResponseRecorder) Pusher() http.Pusher {
	return nil
}

func (r *ginResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, fmt.Errorf("hijack not supported")
}

func (r *ginResponseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseRecorder
}
