package xhttp

import (
	"MiSwap/base/pkg/errcode"
	"context"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

type Response struct {
	TraceId string      `json:"trace_id" example:"a1b2c3d4e5f6g7h8" extensions:"x-order=000"` // 链路追踪id
	Code    uint32      `json:"code" example:"200" extensions:"x-order=001"`                  // 状态码
	Msg     string      `json:"msg" example:"OK" extensions:"x-order=002"`                    // 消息
	Data    interface{} `json:"data" extensions:"x-order=003"`                                // 数据
}

// GetTraceId 获取链路追踪id
func GetTraceId(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}
	//return strings.Replace(uuid.NewString(), "-", "", -1)

	return ""
}

// OkJson 成功json响应返回
func OkJson(c *gin.Context, v interface{}) {
	c.JSON(http.StatusOK, &Response{
		TraceId: GetTraceId(c.Request.Context()),
		Code:    errcode.CodeOK,
		Msg:     errcode.MsgOK,
		Data:    v,
	})
}

// Error 错误响应返回
func Error(c *gin.Context, err error) {
	ctx := c.Request.Context()
	e := errcode.ParseErr(err)
	c.JSON(e.HTTPCode(), &Response{
		TraceId: GetTraceId(ctx),
		Code:    e.Code(),
		Msg:     e.Error(),
		Data:    nil,
	})
}
