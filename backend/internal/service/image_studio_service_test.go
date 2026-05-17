package service

import (
	"context"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
)

func TestExtractImageStudioPromptText_IgnoresResponsesTextConfig(t *testing.T) {
	body := []byte(`{
		"id":"resp_123",
		"model":"gpt-5.4-mini",
		"text":{"format":{"type":"text"},"verbosity":"medium"},
		"output":[
			{
				"type":"message",
				"role":"assistant",
				"content":[
					{"type":"output_text","text":"80年代中国高中毕业照，胶片质感，整齐校服，教学楼前合影"}
				]
			}
		]
	}`)

	got := extractImageStudioPromptText(body)

	if got != "80年代中国高中毕业照，胶片质感，整齐校服，教学楼前合影" {
		t.Fatalf("extractImageStudioPromptText()=%q", got)
	}
}

func TestExtractImageStudioPromptText_PrefersOutputText(t *testing.T) {
	body := []byte(`{
		"text":{"format":{"type":"text"},"verbosity":"medium"},
		"output_text":"optimized prompt"
	}`)

	got := extractImageStudioPromptText(body)

	if got != "optimized prompt" {
		t.Fatalf("extractImageStudioPromptText()=%q", got)
	}
}

func TestExtractImageStudioOutputs_ResponsesResultField(t *testing.T) {
	body := []byte(`{
		"output":[
			{"type":"image_generation_call","result":"aGVsbG8=","revised_prompt":"draw a cat","output_format":"png"}
		]
	}`)

	outputs, err := extractImageStudioOutputs(context.Background(), body)

	if err != nil {
		t.Fatalf("extractImageStudioOutputs() error=%v", err)
	}
	if len(outputs) != 1 {
		t.Fatalf("extractImageStudioOutputs() len=%d", len(outputs))
	}
	if string(outputs[0].Data) != "hello" {
		t.Fatalf("output data=%q", string(outputs[0].Data))
	}
	if outputs[0].RevisedPrompt == nil || *outputs[0].RevisedPrompt != "draw a cat" {
		t.Fatalf("revised prompt=%v", outputs[0].RevisedPrompt)
	}
}

func TestExtractImageStudioOutputs_IgnoresNonImageResultField(t *testing.T) {
	body := []byte(`{
		"output":[
			{"type":"message","result":"aGVsbG8="}
		]
	}`)

	outputs, err := extractImageStudioOutputs(context.Background(), body)

	if err != nil {
		t.Fatalf("extractImageStudioOutputs() error=%v", err)
	}
	if len(outputs) != 0 {
		t.Fatalf("extractImageStudioOutputs() len=%d", len(outputs))
	}
}

func TestBuildImageStudioPromptOptimizationRequestBody_IsLightweight(t *testing.T) {
	body, err := buildImageStudioPromptOptimizationRequestBody(ImageStudioOptimizePromptInput{
		Prompt:     "80年代中国高中毕业合影，一群中国高中生在校园里拍摄正式毕业照，男女生穿着那个年代典型的朴素校服与学生装，白衬衫、蓝色外套、简洁中山装风格服饰，发型符合80年代中国学生特征，神情认真、克制、朴实，采用前排坐坐姿、后排站立的整齐队形，人物居中对称排列，画面构图端正完整，毕业照氛围庄重而怀旧；背景为80年代中国校园环境，老式教学楼、红砖墙、黑板报、操场或校门口，带有时代感的校园细节；纪实摄影风格，真实自然光，柔和日光，轻微胶片颗粒，老照片泛黄色调，清晰锐利，细节丰富，真实感强，温暖怀旧，高质量高清",
		Ratio:      "1:1",
		Resolution: "1K",
		Quality:    "high",
	}, defaultImageStudioPromptModel, "1024x1024")

	if err != nil {
		t.Fatalf("buildImageStudioPromptOptimizationRequestBody() error=%v", err)
	}
	if got := gjson.GetBytes(body, "model").String(); got != "gpt-5.5" {
		t.Fatalf("model=%q", got)
	}
	if got := gjson.GetBytes(body, "reasoning.effort").String(); got != "low" {
		t.Fatalf("reasoning.effort=%q", got)
	}
	if got := gjson.GetBytes(body, "text.verbosity").String(); got != "low" {
		t.Fatalf("text.verbosity=%q", got)
	}
	if got := gjson.GetBytes(body, "max_output_tokens").Int(); got != imageStudioPromptMaxOutputTokens {
		t.Fatalf("max_output_tokens=%d", got)
	}
	inputText := gjson.GetBytes(body, "input.0.content.0.text").String()
	if !gjson.GetBytes(body, "input.0.content.0.text").Exists() || inputText == "" {
		t.Fatalf("input text missing: %s", string(body))
	}
	if !strings.Contains(inputText, "整理、去重、压缩") {
		t.Fatalf("long prompt compression instruction missing: %q", inputText)
	}
}
