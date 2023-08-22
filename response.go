package gogin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/haier-interx/e"
)

const (
	HeaderHDSRequest      = "X-GATEWAY" // HDS网关请求标识
	HeaderHDSRequestValue = HeaderHDSRequest
)

type Response struct {
	*BaseResponse
	Data interface{} `json:"data,omitempty"`
}

type BaseResponse struct {
	Code    int    `json:"code" example:"10000"` // 10000:成功 <br> 10004: 未发现 <br> 15000: 内部错误 <br> https://github.com/haier-interx/e
	Message string `json:"msg" example:"success"`
	Detail  string `json:"detail,omitempty" example:""`
}

func NewResponse(data []byte) (*Response, error) {
	r := new(Response)
	err := json.Unmarshal(data, r)
	return r, err
}

func NewSucResponse(data interface{}) *Response {
	return &Response{&BaseResponse{e.COMMON_SUC.Code, e.COMMON_SUC.Msg, ""}, data}
}

func NewErrResponse(err error, detail string) *Response {
	if detail == "" {
		detail = err.Error()
	} else {
		detail = fmt.Sprintf("%s: %s", detail, err.Error())
	}

	switch v := err.(type) {
	case *e.Err:
		return &Response{&BaseResponse{v.Code, v.Msg, detail}, nil}
	}

	// context cancel
	if err == context.DeadlineExceeded {
		return &Response{&BaseResponse{e.COMMON_INTERNAL_BADGATEWAY.Code, e.COMMON_INTERNAL_BADGATEWAY.Msg, detail}, nil}
	}

	parentErr := errors.Unwrap(err)
	switch v := parentErr.(type) {
	case *e.Err:
		return &Response{&BaseResponse{v.Code, v.Msg, detail}, nil}
	}

	v := e.COMMON_INTERNAL_ERR
	return &Response{&BaseResponse{v.Code, v.Msg, detail}, nil}
}

func SendData(ctx *gin.Context, v interface{}) {
	SendResp(ctx, NewSucResponse(v))
}

func SendErrResp(ctx *gin.Context, err error, detail string) {
	SendResp(ctx, NewErrResponse(err, detail))
}

func SendResp(ctx *gin.Context, resp *Response) {
	if resp.Code != e.COMMON_SUC.Code {
		if resp.Detail != "" {
			PushCtxErr(ctx, fmt.Errorf("%d :: %s :: %s", resp.Code, resp.Message, resp.Detail))
		} else {
			PushCtxErr(ctx, fmt.Errorf("%d :: %s", resp.Code, resp.Message))
		}
		ctx.Abort()
	}
	ctx.Set("Response", resp)

	// 1、不是从HDS网关过来的请求,响应不变
	// 2、成功请求,响应不变
	if !requestFromHdsGateway(ctx) || resp.Code == e.COMMON_SUC.Code {
		if GetQueryParamBool(ctx, "pretty") || GetQueryParamBool(ctx, "debug") {
			ctx.IndentedJSON(200, resp)
		} else {
			ctx.JSON(200, resp)
		}
		return
	}
	// http code 重写
	httpCode := respBodyCodeToHttpStatusCode(resp.Code)
	if GetQueryParamBool(ctx, "pretty") || GetQueryParamBool(ctx, "debug") {
		ctx.IndentedJSON(httpCode, resp)
	} else {
		ctx.JSON(httpCode, resp)
	}
	return
}

func PushCtxErr(ctx *gin.Context, err error) {
	ctx.Error(gin.Error{Err: err, Type: gin.ErrorTypePrivate})
}

// 判断请求是不是从HDS网关过来的
func requestFromHdsGateway(ctx *gin.Context) bool {
	h := ctx.GetHeader(HeaderHDSRequest)
	if h == HeaderHDSRequestValue {
		return true
	}
	return false
}

// RequestIsFromHdsGateway 判断请求是不是从HDS网关过来的
func RequestIsFromHdsGateway(ctx *gin.Context) bool {
	return requestFromHdsGateway(ctx)
}

// 根据响应body的code转换成对应的http code
func respBodyCodeToHttpStatusCode(respBodyCode int) int {
	switch respBodyCode {
	case e.COMMON_BADREQUEST.Code,
		e.COMMON_PARAM_ERR.Code,
		e.COMMON_PARAM_MISS.Code:
		return http.StatusBadRequest
	case e.COMMON_NOT_FOUND.Code:
		return http.StatusNotFound
	case e.COMMON_CONFILCT.Code:
		return http.StatusConflict
	case e.COMMON_INTERNAL_ERR.Code,
		e.COMMON_INTERNAL_CALLING_ERR.Code,
		e.COMMON_INTERNAL_CALLING_TIMEOUT.Code,
		e.COMMON_INTERNAL_DB_ERR.Code,
		e.COMMON_INTERNAL_BADGATEWAY.Code:
		return http.StatusInternalServerError
	case e.AUTH_NOTLOGIN.Code:
		return http.StatusUnauthorized
	case e.AUTH_NOPERMISSION.Code,
		e.AUTHINTERNAL_ROLE_NOT_DEFINED.Code,
		e.AUTHINTERNAL_ROLE_NOT_SUPPORT.Code,
		e.AUTHINTERNAL_ROLE_INVALID.Code:
		return http.StatusForbidden
	default:
		return http.StatusOK
	}
}
