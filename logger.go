package gogin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

// defaultLogFormatter is the default log format function Logger middleware uses.
var defaultLogFormatter = func(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		// Truncate in a golang < 1.8 safe way
		param.Latency = param.Latency - param.Latency%time.Second
	}

	var respErrCode = -1
	var respErrMsg string
	var respDetail string
	if v, found := param.Keys["Response"]; found {
		if resp, ok := v.(*Response); ok {
			respErrCode = resp.Code
			respErrMsg = resp.Message
			respDetail = resp.Detail
		}
	}

	return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %s %d '%s' '%s'\n%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		respErrCode, respErrMsg, respDetail,
		param.ErrorMessage,
	)
}

func Logger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{Formatter: defaultLogFormatter})
}
