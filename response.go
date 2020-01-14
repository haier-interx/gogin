package gogin

import (
	"encoding/json"
	"errors"
	"fmt"

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

	parent_err := errors.Unwrap(err)
	switch v := parent_err.(type) {
	case *e.Err:
		return &Response{&BaseResponse{v.Code, v.Msg, detail}, nil}
	}

	v := e.COMMON_INTERNAL_ERR
	return &Response{&BaseResponse{v.Code, v.Msg, detail}, nil}
}
