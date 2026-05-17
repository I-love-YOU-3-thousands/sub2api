package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

func NewImageStudioStorage(cfg *config.Config) (ImageStudioStorage, error) {
	storageType := ImageStudioStorageDriverLocal
	if cfg != nil && strings.TrimSpace(cfg.ImageStudio.Storage.Type) != "" {
		storageType = strings.ToLower(strings.TrimSpace(cfg.ImageStudio.Storage.Type))
	}
	switch storageType {
	case ImageStudioStorageDriverLocal:
		return newLocalImageStudioStorage(cfg), nil
	case ImageStudioStorageDriverS3:
		return newS3ImageStudioStorage(context.Background(), cfg)
	default:
		return nil, fmt.Errorf("unsupported image studio storage type %q", storageType)
	}
}

type localImageStudioStorage struct {
	root string
}

func newLocalImageStudioStorage(cfg *config.Config) *localImageStudioStorage {
	root := ""
	if cfg != nil {
		root = strings.TrimSpace(cfg.ImageStudio.Storage.LocalPath)
	}
	if root == "" {
		root = strings.TrimSpace(os.Getenv("DATA_DIR"))
	}
	if root == "" && cfg != nil {
		root = strings.TrimSpace(cfg.Pricing.DataDir)
	}
	if root == "" {
		root = "./data"
	}
	if !strings.HasSuffix(filepath.Clean(root), "image-studio") {
		root = filepath.Join(root, "image-studio")
	}
	return &localImageStudioStorage{root: root}
}

func (s *localImageStudioStorage) Save(ctx context.Context, input ImageStudioStoreInput) (*ImageStudioStoredAsset, error) {
	if len(input.Data) == 0 {
		return nil, fmt.Errorf("image data is empty")
	}
	key := buildImageStudioStorageKey(input, "")
	path := filepath.Join(s.root, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create image studio cache dir: %w", err)
	}
	if err := os.WriteFile(path, input.Data, 0o644); err != nil {
		return nil, fmt.Errorf("write image studio asset: %w", err)
	}
	width, height := decodeImageDimensions(input.Data)
	return &ImageStudioStoredAsset{
		StorageDriver: ImageStudioStorageDriverLocal,
		StorageKey:    key,
		MimeType:      normalizeImageStudioMimeType(input.MimeType, input.FileName, input.Data),
		Width:         width,
		Height:        height,
		SizeBytes:     int64(len(input.Data)),
	}, nil
}

func (s *localImageStudioStorage) Open(ctx context.Context, asset *ImageGenerationAsset) (io.ReadCloser, string, int64, error) {
	if asset == nil {
		return nil, "", 0, fmt.Errorf("asset is nil")
	}
	clean := filepath.Clean(filepath.FromSlash(asset.StorageKey))
	path := filepath.Join(s.root, clean)
	root, err := filepath.Abs(s.root)
	if err != nil {
		return nil, "", 0, err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, "", 0, err
	}
	if abs != root && !strings.HasPrefix(abs, root+string(os.PathSeparator)) {
		return nil, "", 0, fmt.Errorf("invalid image studio storage key")
	}
	file, err := os.Open(abs)
	if err != nil {
		return nil, "", 0, err
	}
	return file, asset.MimeType, asset.SizeBytes, nil
}

type s3ImageStudioStorage struct {
	client *s3.Client
	bucket string
	prefix string
}

func newS3ImageStudioStorage(ctx context.Context, cfg *config.Config) (*s3ImageStudioStorage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("image studio s3 config is required")
	}
	scfg := cfg.ImageStudio.Storage
	region := strings.TrimSpace(scfg.Region)
	if region == "" {
		region = "auto"
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(scfg.AccessKeyID, scfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load image studio s3 config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if scfg.Endpoint != "" {
			o.BaseEndpoint = &scfg.Endpoint
		}
		if scfg.ForcePathStyle {
			o.UsePathStyle = true
		}
		o.APIOptions = append(o.APIOptions, v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware)
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
	})
	prefix := strings.Trim(strings.TrimSpace(scfg.Prefix), "/")
	return &s3ImageStudioStorage{client: client, bucket: scfg.Bucket, prefix: prefix}, nil
}

func (s *s3ImageStudioStorage) Save(ctx context.Context, input ImageStudioStoreInput) (*ImageStudioStoredAsset, error) {
	if len(input.Data) == 0 {
		return nil, fmt.Errorf("image data is empty")
	}
	key := buildImageStudioStorageKey(input, s.prefix)
	contentType := normalizeImageStudioMimeType(input.MimeType, input.FileName, input.Data)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &key,
		Body:        bytes.NewReader(input.Data),
		ContentType: &contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("image studio s3 put object: %w", err)
	}
	width, height := decodeImageDimensions(input.Data)
	return &ImageStudioStoredAsset{
		StorageDriver: ImageStudioStorageDriverS3,
		StorageKey:    key,
		MimeType:      contentType,
		Width:         width,
		Height:        height,
		SizeBytes:     int64(len(input.Data)),
	}, nil
}

func (s *s3ImageStudioStorage) Open(ctx context.Context, asset *ImageGenerationAsset) (io.ReadCloser, string, int64, error) {
	if asset == nil {
		return nil, "", 0, fmt.Errorf("asset is nil")
	}
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &asset.StorageKey,
	})
	if err != nil {
		return nil, "", 0, fmt.Errorf("image studio s3 get object: %w", err)
	}
	contentType := asset.MimeType
	if result.ContentType != nil && strings.TrimSpace(*result.ContentType) != "" {
		contentType = *result.ContentType
	}
	return result.Body, contentType, asset.SizeBytes, nil
}

func buildImageStudioStorageKey(input ImageStudioStoreInput, prefix string) string {
	now := time.Now()
	ext := imageStudioExtension(input.MimeType, input.FileName, input.Data)
	taskID := strings.TrimSpace(input.TaskID)
	if taskID == "" {
		taskID = uuid.NewString()
	}
	name := fmt.Sprintf("%s_%s_%02d_%s%s", taskID, input.Kind, input.Seq, strings.ReplaceAll(uuid.NewString(), "-", ""), ext)
	parts := []string{}
	if prefix != "" {
		parts = append(parts, prefix)
	}
	parts = append(parts, now.Format("2006"), now.Format("01"), now.Format("02"), name)
	return strings.Join(parts, "/")
}

func normalizeImageStudioMimeType(mimeType, fileName string, data []byte) string {
	if media, _, err := mime.ParseMediaType(strings.TrimSpace(mimeType)); err == nil && strings.HasPrefix(media, "image/") {
		return media
	}
	if ext := strings.ToLower(filepath.Ext(fileName)); ext != "" {
		if detected := mime.TypeByExtension(ext); strings.HasPrefix(detected, "image/") {
			if media, _, err := mime.ParseMediaType(detected); err == nil {
				return media
			}
		}
	}
	if len(data) > 0 {
		if detected := http.DetectContentType(data); strings.HasPrefix(detected, "image/") {
			return detected
		}
	}
	return "image/png"
}

func imageStudioExtension(mimeType, fileName string, data []byte) string {
	if ext := strings.ToLower(filepath.Ext(fileName)); ext != "" && len(ext) <= 8 {
		return ext
	}
	switch normalizeImageStudioMimeType(mimeType, fileName, data) {
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".png"
	}
}

func decodeImageDimensions(data []byte) (int, int) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}
