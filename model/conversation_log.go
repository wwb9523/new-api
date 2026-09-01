package model

import (
	"context"
	"errors"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

const (
	ConversationLogStatusSuccess = "success"
	ConversationLogStatusError   = "error"
)

type ConversationLog struct {
	Id               int    `json:"id" gorm:"index:idx_conversation_logs_created_at_id,priority:2;index:idx_conversation_logs_user_id_id,priority:2"`
	RequestId        string `json:"request_id" gorm:"type:varchar(64);index:idx_conversation_logs_request_id;default:''"`
	UserId           int    `json:"user_id" gorm:"index;index:idx_conversation_logs_user_id_id,priority:1"`
	Username         string `json:"username" gorm:"index;index:idx_conversation_logs_username_model,priority:2;default:''"`
	CreatedAt        int64  `json:"created_at" gorm:"bigint;index:idx_conversation_logs_created_at_id,priority:1"`
	Status           string `json:"status" gorm:"type:varchar(16);index;default:''"`
	RelayFormat      string `json:"relay_format" gorm:"type:varchar(32);index;default:''"`
	RelayMode        int    `json:"relay_mode" gorm:"index;default:0"`
	ModelName        string `json:"model_name" gorm:"index;index:idx_conversation_logs_username_model,priority:1;default:''"`
	TokenName        string `json:"token_name" gorm:"index;default:''"`
	TokenId          int    `json:"token_id" gorm:"default:0;index"`
	ChannelId        int    `json:"channel" gorm:"column:channel_id;index"`
	ChannelName      string `json:"channel_name" gorm:"->"`
	Group            string `json:"group" gorm:"index"`
	PromptTokens     int    `json:"prompt_tokens" gorm:"default:0"`
	CompletionTokens int    `json:"completion_tokens" gorm:"default:0"`
	Quota            int    `json:"quota" gorm:"default:0"`
	UseTime          int    `json:"use_time" gorm:"default:0"`
	IsStream         bool   `json:"is_stream"`
	RequestPreview   string `json:"request_preview"`
	ResponsePreview  string `json:"response_preview"`
	RequestBody      string `json:"request_body"`
	ResponseBody     string `json:"response_body"`
	ErrorMessage     string `json:"error_message"`
	Truncated        bool   `json:"truncated"`
}

type RecordConversationLogParams struct {
	RequestId        string
	UserId           int
	Username         string
	Status           string
	RelayFormat      string
	RelayMode        int
	ModelName        string
	TokenName        string
	TokenId          int
	ChannelId        int
	Group            string
	PromptTokens     int
	CompletionTokens int
	Quota            int
	UseTimeSeconds   int
	IsStream         bool
	RequestPreview   string
	ResponsePreview  string
	RequestBody      string
	ResponseBody     string
	ErrorMessage     string
	Truncated        bool
}

func RecordConversationLog(params RecordConversationLogParams) {
	if !common.ConversationLogEnabled {
		return
	}
	if params.RequestId == "" {
		params.RequestId = common.NewRequestId()
	}
	if params.Status == "" {
		params.Status = ConversationLogStatusSuccess
	}
	log := &ConversationLog{
		RequestId:        params.RequestId,
		UserId:           params.UserId,
		Username:         params.Username,
		CreatedAt:        common.GetTimestamp(),
		Status:           params.Status,
		RelayFormat:      params.RelayFormat,
		RelayMode:        params.RelayMode,
		ModelName:        params.ModelName,
		TokenName:        params.TokenName,
		TokenId:          params.TokenId,
		ChannelId:        params.ChannelId,
		Group:            params.Group,
		PromptTokens:     params.PromptTokens,
		CompletionTokens: params.CompletionTokens,
		Quota:            params.Quota,
		UseTime:          params.UseTimeSeconds,
		IsStream:         params.IsStream,
		RequestPreview:   params.RequestPreview,
		ResponsePreview:  params.ResponsePreview,
		RequestBody:      params.RequestBody,
		ResponseBody:     params.ResponseBody,
		ErrorMessage:     params.ErrorMessage,
		Truncated:        params.Truncated,
	}
	if err := LOG_DB.Create(log).Error; err != nil {
		common.SysLog("failed to record conversation log: " + err.Error())
	}
}

func GetAllConversationLogs(startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, startIdx int, num int, channel int, group string, requestId string, status string) (logs []*ConversationLog, total int64, err error) {
	tx := LOG_DB.Model(&ConversationLog{})
	tx, err = applyConversationLogFilters(tx, startTimestamp, endTimestamp, modelName, username, tokenName, channel, group, requestId, status)
	if err != nil {
		return nil, 0, err
	}
	if err = tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := "conversation_logs.created_at desc, conversation_logs.id desc"
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		order = "conversation_logs.created_at desc, conversation_logs.request_id desc"
	}
	if err = tx.Order(order).Limit(num).Offset(startIdx).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		assignConversationLogDisplayIds(logs, startIdx)
	}
	fillConversationLogChannelNames(logs)
	return logs, total, nil
}

func GetUserConversationLogs(userId int, startTimestamp int64, endTimestamp int64, modelName string, tokenName string, startIdx int, num int, group string, requestId string, status string) (logs []*ConversationLog, total int64, err error) {
	tx := LOG_DB.Model(&ConversationLog{}).Where("conversation_logs.user_id = ?", userId)
	tx, err = applyConversationLogFilters(tx, startTimestamp, endTimestamp, modelName, "", tokenName, 0, group, requestId, status)
	if err != nil {
		return nil, 0, err
	}
	if err = tx.Limit(logSearchCountLimit).Count(&total).Error; err != nil {
		common.SysError("failed to count user conversation logs: " + err.Error())
		return nil, 0, errors.New("查询对话日志失败")
	}
	order := "conversation_logs.id desc"
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		order = "conversation_logs.created_at desc, conversation_logs.request_id desc"
	}
	if err = tx.Order(order).Limit(num).Offset(startIdx).Find(&logs).Error; err != nil {
		common.SysError("failed to search user conversation logs: " + err.Error())
		return nil, 0, errors.New("查询对话日志失败")
	}
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		assignConversationLogDisplayIds(logs, startIdx)
	}
	return logs, total, nil
}

func CountOldConversationLog(ctx context.Context, targetTimestamp int64) (int64, error) {
	var total int64
	if err := LOG_DB.WithContext(ctx).Model(&ConversationLog{}).Where("created_at < ?", targetTimestamp).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func DeleteOldConversationLogBatch(ctx context.Context, targetTimestamp int64, limit int) (int64, error) {
	if targetTimestamp <= 0 {
		return 0, errors.New("target timestamp must be positive")
	}
	if limit <= 0 {
		limit = 100
	}
	if nil != ctx.Err() {
		return 0, ctx.Err()
	}
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		total, err := CountOldConversationLog(ctx, targetTimestamp)
		if err != nil {
			return 0, err
		}
		if total == 0 {
			return 0, nil
		}
		if err := LOG_DB.WithContext(ctx).Exec(
			"ALTER TABLE conversation_logs DELETE WHERE created_at < ? SETTINGS mutations_sync = 1",
			targetTimestamp,
		).Error; err != nil {
			return 0, err
		}
		return total, nil
	}
	result := LOG_DB.WithContext(ctx).Where("created_at < ?", targetTimestamp).Limit(limit).Delete(&ConversationLog{})
	return result.RowsAffected, result.Error
}

func applyConversationLogFilters(tx *gorm.DB, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, channel int, group string, requestId string, status string) (*gorm.DB, error) {
	var err error
	if tx, err = applyExplicitLogTextFilter(tx, "conversation_logs.model_name", modelName); err != nil {
		return nil, err
	}
	if tx, err = applyExplicitLogTextFilter(tx, "conversation_logs.username", username); err != nil {
		return nil, err
	}
	if tokenName != "" {
		tx = tx.Where("conversation_logs.token_name = ?", tokenName)
	}
	if requestId != "" {
		tx = tx.Where("conversation_logs.request_id = ?", requestId)
	}
	if status != "" {
		tx = tx.Where("conversation_logs.status = ?", status)
	}
	if startTimestamp != 0 {
		tx = tx.Where("conversation_logs.created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("conversation_logs.created_at <= ?", endTimestamp)
	}
	if channel != 0 {
		tx = tx.Where("conversation_logs.channel_id = ?", channel)
	}
	if group != "" {
		tx = tx.Where("conversation_logs."+logGroupCol+" = ?", group)
	}
	return tx, nil
}

func assignConversationLogDisplayIds(logs []*ConversationLog, startIdx int) {
	for i := range logs {
		logs[i].Id = startIdx + i + 1
	}
}

func fillConversationLogChannelNames(logs []*ConversationLog) {
	channelIds := make([]int, 0)
	seen := make(map[int]struct{})
	for _, log := range logs {
		if log.ChannelId == 0 {
			continue
		}
		if _, ok := seen[log.ChannelId]; ok {
			continue
		}
		seen[log.ChannelId] = struct{}{}
		channelIds = append(channelIds, log.ChannelId)
	}
	if len(channelIds) == 0 {
		return
	}

	var channels []struct {
		Id   int    `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}
	if common.MemoryCacheEnabled {
		for _, channelId := range channelIds {
			if cacheChannel, err := CacheGetChannel(channelId); err == nil {
				channels = append(channels, struct {
					Id   int    `gorm:"column:id"`
					Name string `gorm:"column:name"`
				}{Id: channelId, Name: cacheChannel.Name})
			}
		}
	} else if err := DB.Table("channels").Select("id, name").Where("id IN ?", channelIds).Find(&channels).Error; err != nil {
		return
	}
	channelMap := make(map[int]string, len(channels))
	for _, channel := range channels {
		channelMap[channel.Id] = channel.Name
	}
	for i := range logs {
		logs[i].ChannelName = channelMap[logs[i].ChannelId]
	}
}
