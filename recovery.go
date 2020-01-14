package gogin

import (
	"github.com/gin-gonic/gin"
	"github.com/haier-interx/e"
	"net/http"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		gin.Recovery()(c)

		// panic 返回正常格式结果
		if c.IsAborted() && c.Writer.Status() == http.StatusInternalServerError {
			c.JSON(200, NewErrResponse(e.COMMON_INTERNAL_ERR, "[Omg! panic!]"))
		}
	}
}
