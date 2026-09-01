package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

func GetAllConversationLogs(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	channel, _ := strconv.Atoi(c.Query("channel"))
	logs, total, err := model.GetAllConversationLogs(
		startTimestamp,
		endTimestamp,
		c.Query("model_name"),
		c.Query("username"),
		c.Query("token_name"),
		pageInfo.GetStartIdx(),
		pageInfo.GetPageSize(),
		channel,
		c.Query("group"),
		c.Query("request_id"),
		c.Query("status"),
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(logs)
	common.ApiSuccess(c, pageInfo)
}

func GetUserConversationLogs(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	logs, total, err := model.GetUserConversationLogs(
		c.GetInt("id"),
		startTimestamp,
		endTimestamp,
		c.Query("model_name"),
		c.Query("token_name"),
		pageInfo.GetStartIdx(),
		pageInfo.GetPageSize(),
		c.Query("group"),
		c.Query("request_id"),
		c.Query("status"),
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(logs)
	common.ApiSuccess(c, pageInfo)
}
