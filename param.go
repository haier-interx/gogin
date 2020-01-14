package gogin

import "github.com/gin-gonic/gin"

func GetQueryParamBool(ctx *gin.Context, name string) bool {
	str, found := ctx.GetQuery(name)
	if found {
		switch str {
		case "":
			return true
		case "0", "false", "False":
			return false
		default:
			return true
		}
	} else {
		return false
	}
}
func SetParam(ctx *gin.Context, name, value string) {
	for i, entry := range ctx.Params {
		if entry.Key == name {
			ctx.Params[i].Value = value
			return
		}
	}
	ps := make([]gin.Param, len(ctx.Params)+1)
	copy(ps, ctx.Params)
	ps[len(ctx.Params)] = gin.Param{name, value}
	ctx.Params = ps
}
