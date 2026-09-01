package service

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"

	"github.com/gin-gonic/gin"
)

const conversationPreviewRunes = 240
const conversationRequestBodyKey = "conversation_log_request_body"
const conversationResponseBodyKey = "conversation_log_response_body"
const conversationTruncatedKey = "conversation_log_truncated"

func BeginConversationLog(c *gin.Context, relayInfo *relaycommon.RelayInfo) {
	if !common.ConversationLogEnabled || c == nil || relayInfo == nil {
		return
	}
	if relayInfo.ConversationLog == nil {
		relayInfo.ConversationLog = &relaycommon.ConversationLogInfo{}
	}
	if storage, err := common.GetBodyStorage(c); err == nil {
		if body, bodyErr := storage.Bytes(); bodyErr == nil {
			relayInfo.ConversationLog.RequestBody, relayInfo.ConversationLog.Truncated = truncateConversationText(string(body), false)
			c.Set(conversationRequestBodyKey, relayInfo.ConversationLog.RequestBody)
			c.Set(conversationTruncatedKey, relayInfo.ConversationLog.Truncated)
		}
	}
}

func AppendConversationResponseChunk(c *gin.Context, text string) {
	if !common.ConversationLogEnabled || c == nil || text == "" || text == "[DONE]" {
		return
	}
	current := c.GetString(conversationResponseBodyKey)
	combined := current + text
	truncated, didTruncate := truncateConversationText(combined, c.GetBool(conversationTruncatedKey))
	c.Set(conversationResponseBodyKey, truncated)
	c.Set(conversationTruncatedKey, didTruncate)
}

func AppendConversationResponseText(relayInfo *relaycommon.RelayInfo, text string) {
	if !common.ConversationLogEnabled || relayInfo == nil || text == "" {
		return
	}
	if relayInfo.ConversationLog == nil {
		relayInfo.ConversationLog = &relaycommon.ConversationLogInfo{}
	}
	combined := relayInfo.ConversationLog.ResponseBody + text
	var truncated bool
	relayInfo.ConversationLog.ResponseBody, truncated = truncateConversationText(combined, relayInfo.ConversationLog.Truncated)
	relayInfo.ConversationLog.Truncated = truncated
	relayInfo.ConversationLog.ResponsePreview = previewText(relayInfo.ConversationLog.ResponseBody)
}

func SetConversationResponseBody(c *gin.Context, relayInfo *relaycommon.RelayInfo, body string) {
	if !common.ConversationLogEnabled || relayInfo == nil {
		return
	}
	if relayInfo.ConversationLog == nil {
		relayInfo.ConversationLog = &relaycommon.ConversationLogInfo{}
	}
	var truncated bool
	relayInfo.ConversationLog.ResponseBody, truncated = truncateConversationText(body, relayInfo.ConversationLog.Truncated)
	relayInfo.ConversationLog.Truncated = truncated
	relayInfo.ConversationLog.ResponsePreview = previewText(relayInfo.ConversationLog.ResponseBody)
	if c != nil {
		c.Set(conversationResponseBodyKey, relayInfo.ConversationLog.ResponseBody)
		c.Set(conversationTruncatedKey, relayInfo.ConversationLog.Truncated)
	}
}

func SetFinalConversationUsage(relayInfo *relaycommon.RelayInfo, usage *dto.Usage, quota int) {
	if !common.ConversationLogEnabled || relayInfo == nil {
		return
	}
	if usage == nil {
		SetFinalConversationTokenUsage(relayInfo, 0, 0, quota)
		return
	}
	SetFinalConversationTokenUsage(relayInfo, usage.PromptTokens, usage.CompletionTokens, quota)
}

func SetFinalConversationTokenUsage(relayInfo *relaycommon.RelayInfo, promptTokens int, completionTokens int, quota int) {
	if !common.ConversationLogEnabled || relayInfo == nil {
		return
	}
	if relayInfo.ConversationLog == nil {
		relayInfo.ConversationLog = &relaycommon.ConversationLogInfo{}
	}
	relayInfo.ConversationLog.PromptTokens = promptTokens
	relayInfo.ConversationLog.CompletionTokens = completionTokens
	relayInfo.ConversationLog.Quota = quota
}

func RecordFinalConversationLog(c *gin.Context, relayInfo *relaycommon.RelayInfo, usage *dto.Usage, quota int, apiErr error) {
	if !common.ConversationLogEnabled || c == nil || relayInfo == nil {
		return
	}
	conv := relayInfo.ConversationLog
	if conv == nil {
		conv = &relaycommon.ConversationLogInfo{}
	}
	if conv.RequestBody == "" {
		conv.RequestBody = c.GetString(conversationRequestBodyKey)
	}
	if conv.ResponseBody == "" {
		conv.ResponseBody = c.GetString(conversationResponseBodyKey)
		conv.ResponsePreview = previewText(conv.ResponseBody)
	}
	conv.Truncated = conv.Truncated || c.GetBool(conversationTruncatedKey)
	status := model.ConversationLogStatusSuccess
	errorMessage := ""
	if apiErr != nil {
		status = model.ConversationLogStatusError
		errorMessage = apiErr.Error()
	}
	promptTokens := 0
	completionTokens := 0
	if usage != nil {
		promptTokens = usage.PromptTokens
		completionTokens = usage.CompletionTokens
	} else {
		promptTokens = conv.PromptTokens
		completionTokens = conv.CompletionTokens
	}
	if quota == 0 {
		quota = conv.Quota
	}
	useTimeSeconds := int(time.Since(relayInfo.StartTime).Seconds())
	if useTimeSeconds < 0 {
		useTimeSeconds = 0
	}
	model.RecordConversationLog(model.RecordConversationLogParams{
		RequestId:        relayInfo.RequestId,
		UserId:           relayInfo.UserId,
		Username:         c.GetString("username"),
		Status:           status,
		RelayFormat:      string(relayInfo.RelayFormat),
		RelayMode:        relayInfo.RelayMode,
		ModelName:        relayInfo.OriginModelName,
		TokenName:        c.GetString("token_name"),
		TokenId:          relayInfo.TokenId,
		ChannelId:        relayInfo.ChannelId,
		Group:            relayInfo.UsingGroup,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		Quota:            quota,
		UseTimeSeconds:   useTimeSeconds,
		IsStream:         relayInfo.IsStream,
		RequestPreview:   previewText(conv.RequestBody),
		ResponsePreview:  conv.ResponsePreview,
		RequestBody:      conv.RequestBody,
		ResponseBody:     conv.ResponseBody,
		ErrorMessage:     errorMessage,
		Truncated:        conv.Truncated,
	})
}

func ConversationResponseTextFromOpenAIResponse(resp *dto.OpenAITextResponse) string {
	if resp == nil {
		return ""
	}
	var builder strings.Builder
	for _, choice := range resp.Choices {
		if content := choice.Message.StringContent(); content != "" {
			builder.WriteString(content)
		}
		if reasoning := choice.Message.GetReasoningContent(); reasoning != "" {
			builder.WriteString(reasoning)
		}
	}
	return builder.String()
}

func ConversationResponseTextFromResponsesResponse(resp *dto.OpenAIResponsesResponse) string {
	return ExtractOutputTextFromResponses(resp)
}

func ConversationResponseTextFromClaudeResponse(resp *dto.ClaudeResponse) string {
	if resp == nil {
		return ""
	}
	var builder strings.Builder
	for _, block := range resp.Content {
		if block.Text != nil && *block.Text != "" {
			builder.WriteString(*block.Text)
		}
		if block.Thinking != nil && *block.Thinking != "" {
			builder.WriteString(*block.Thinking)
		}
	}
	if resp.Completion != "" {
		builder.WriteString(resp.Completion)
	}
	return builder.String()
}

func truncateConversationText(text string, alreadyTruncated bool) (string, bool) {
	maxBytes := common.ConversationLogMaxBodyBytes
	if maxBytes <= 0 {
		return text, alreadyTruncated
	}
	if len(text) <= maxBytes {
		return text, alreadyTruncated
	}
	truncated := text[:maxBytes]
	for !utf8.ValidString(truncated) && len(truncated) > 0 {
		truncated = truncated[:len(truncated)-1]
	}
	return truncated, true
}

func previewText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= conversationPreviewRunes {
		return text
	}
	return string(runes[:conversationPreviewRunes])
}
