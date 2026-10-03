package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
	"github.com/unicornfairy864/LNF-SERVER/global"
)

type Client struct{}

// ==================== 错误归一 ====================
// 上层（service）按这三类做分支：超时/上游失败 → 业务降级；空响应 → 重试或降级。
// 详细原因保留在错误信息中（fmt.Errorf 包裹），便于日志排查。
var (
	ErrTimeout       = errors.New("llm: timeout")        // 调用超时
	ErrUpstream      = errors.New("llm: upstream error") // 上游返回错误（网络/鉴权/限流/参数等）
	ErrEmptyResponse = errors.New("llm: empty response") // 上游 200 但无有效内容
)

const (
	// llmDefaultTimeout 配置缺失时的兜底超时（正常路径读 config agent_timeout）
	llmDefaultTimeout = 30 * time.Second
	// llmMaxTokens 单次生成上限，防御性控制成本（精排 20 条候选也足够）
	llmMaxTokens = 2048
	// llmMaxRetries 失败后的额外重试次数（同一次调用总预算受超时约束）
	llmMaxRetries = 1
)

// EasyRequest 单轮纯文本调用（emoji 谐音翻译等轻量场景沿用，行为与既有版本一致）
func (c *Client) EasyRequest(system, user string) (string, error) {
	ctx := context.Background()

	params := openai.ChatCompletionNewParams{
		Model: global.LNF_CONFIG.OpenAI.DefaultModel,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system),
			openai.UserMessage(user),
		},
		Temperature: openai.Float(global.LNF_CONFIG.OpenAI.Temperature),
	}

	resp, err := global.LNF_OpenAI.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}
	return resp.Choices[0].Message.Content, nil
}

// RequestJSON 结构化输出调用（文本）：强制 JSON object 模式 + 超时 + 1 次重试。
// 返回的是模型原始文本（可能带 Markdown 围栏），调用方需用 schema.ParseJSONObject 解析。
func (c *Client) RequestJSON(system, user string) (string, error) {
	parts := []openai.ChatCompletionContentPartUnionParam{openai.TextContentPart(user)}
	return c.chatJSON(system, parts, "")
}

// RequestJSONVision 结构化输出调用（文本 + 图片）。
// 统一使用 openai.default_model（要求该模型具备视觉能力，如 glm-5.3-flash / deepseek-4.1）；
// images 为空时等价 RequestJSON；若上游因不支持图片而报错，自动降级为纯文本重试一次（记日志降级，不阻断主流程）。
func (c *Client) RequestJSONVision(system, user string, imageURLs []string) (string, error) {
	parts := make([]openai.ChatCompletionContentPartUnionParam, 0, len(imageURLs)+1)
	parts = append(parts, openai.TextContentPart(user))
	for _, u := range imageURLs {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		parts = append(parts, openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{URL: u}))
	}
	if len(parts) == 1 { // 只有文本，退化为普通 JSON 调用
		return c.chatJSON(system, parts, "")
	}
	out, err := c.chatJSON(system, parts, "")
	if err == nil {
		return out, nil
	}
	// 仅当上游错误明显与图片输入相关时降级，避免把限流/超时误判为多模态不可用
	if errors.Is(err, ErrUpstream) && isVisionUnsupported(err) {
		return c.chatJSON(system, []openai.ChatCompletionContentPartUnionParam{openai.TextContentPart(user)}, "")
	}
	return "", err
}

// chatJSON 统一的结构化调用入口
// 流程：JSON mode 请求 → 若上游明确不支持 response_format 则去掉该参数重试一次 → 失败按错误归一返回
func (c *Client) chatJSON(system string, parts []openai.ChatCompletionContentPartUnionParam, model string) (string, error) {
	if model == "" {
		model = global.LNF_CONFIG.OpenAI.DefaultModel
	}
	timeout := global.LNF_CONFIG.OpenAI.AgentValues().Timeout

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	build := func(jsonMode bool) openai.ChatCompletionNewParams {
		params := openai.ChatCompletionNewParams{
			Model: model,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(system),
				openai.UserMessage(parts),
			},
			Temperature: openai.Float(global.LNF_CONFIG.OpenAI.Temperature),
			MaxTokens:   openai.Int(llmMaxTokens),
		}
		if jsonMode {
			rf := shared.NewResponseFormatJSONObjectParam()
			params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{OfJSONObject: &rf}
		}
		return params
	}

	jsonMode := true
	var lastErr error
	for attempt := 0; attempt <= llmMaxRetries; attempt++ {
		resp, err := global.LNF_OpenAI.Chat.Completions.New(ctx, build(jsonMode))
		if err != nil {
			// 上游不支持 JSON mode（400 且提示 response_format）→ 去掉该参数重试，不算一次失败
			if jsonMode && isResponseFormatUnsupported(err) {
				jsonMode = false
				lastErr = normalizeErr(err, ctx)
				continue
			}
			lastErr = normalizeErr(err, ctx)
			if errors.Is(lastErr, ErrTimeout) {
				break
			}
			continue
		}
		if len(resp.Choices) == 0 {
			lastErr = ErrEmptyResponse
			continue
		}
		content := strings.TrimSpace(resp.Choices[0].Message.Content)
		if content == "" {
			lastErr = ErrEmptyResponse
			continue
		}
		return content, nil
	}
	return "", lastErr
}

// normalizeErr 错误归一：区分「超时取消」与「上游错误」，并保留原始信息
func normalizeErr(err error, ctx context.Context) error {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) || ctx.Err() != nil {
		return fmt.Errorf("%w: %v", ErrTimeout, err)
	}
	return fmt.Errorf("%w: %v", ErrUpstream, err)
}

// isResponseFormatUnsupported 判断上游是否因不支持 response_format 而报错
func isResponseFormatUnsupported(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "response_format") || strings.Contains(msg, "json_object")
}

// isVisionUnsupported 判断上游错误是否与图片/视觉输入相关（用于降级为纯文本）
func isVisionUnsupported(err error) bool {
	msg := strings.ToLower(err.Error())
	hints := []string{"image", "vision", "multimodal", "content_part", "image_url"}
	for _, h := range hints {
		if strings.Contains(msg, h) {
			return true
		}
	}
	return false
}
