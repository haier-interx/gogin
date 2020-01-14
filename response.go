package gogin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"

	"github.com/haier-interx/e"
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

	parent_err := errors.Unwrap(err)
	switch v := parent_err.(type) {
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
	if resp.Code != 10000 {
		if resp.Detail != "" {
			PushCtxErr(ctx, fmt.Errorf("%d :: %s :: %s", resp.Code, resp.Message, resp.Detail))
		} else {
			PushCtxErr(ctx, fmt.Errorf("%d :: %s", resp.Code, resp.Message))
		}
		ctx.Abort()
	}

	//if !GetQueryParamBool(ctx, "debug") {
	//	resp.Detail = ""
	//}

	ctx.Set("Response", resp)
	if GetQueryParamBool(ctx, "pretty") || GetQueryParamBool(ctx, "debug") {
		ctx.IndentedJSON(200, resp)
	} else {
		ctx.JSON(200, resp)
	}
}

func PushCtxErr(ctx *gin.Context, err error) {
	ctx.Error(gin.Error{Err: err, Type: gin.ErrorTypePrivate})
}
