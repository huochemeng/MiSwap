package errcode

import (
	"net/http"

	"github.com/pkg/errors"
	"google.golang.org/grpc/status"
)

const (
	// CodeOK 请求成功业务状态码
	CodeOK = 200
	// MsgOK 请求成功消息
	MsgOK = "Successful"

	// CodeCustom 自定义错误业务状态码
	CodeCustom = 7000
	// MsgCustom 自定义错误消息
	MsgCustom = "Custom error"
)

// Err 业务错误结构
type Err struct {
	code     uint32
	httpCode int
	msg      string
}

// Code 业务状态码
func (e *Err) Code() uint32 {
	return e.code
}

// HTTPCode HTTP状态码
func (e *Err) HTTPCode() int {
	return e.httpCode
}

// Error 消息
func (e *Err) Error() string {
	return e.msg
}

// 业务错误
var (
	NoErr                   = NewErr(CodeOK, MsgOK)
	ErrCustom               = NewErr(CodeCustom, MsgCustom)
	ErrUnexpected           = NewErr(7777, "Network error, please try again later", http.StatusInternalServerError)
	ErrTokenNotValidYet     = NewErr(9999, "Token illegal", http.StatusUnauthorized)
	ErrInvalidParams        = NewErr(10002, "Parameter is illegal")
	ErrTokenVerify          = NewErr(10003, "Token check error", http.StatusUnauthorized)
	ErrTokenExpire          = NewErr(10004, "Expired token", http.StatusUnauthorized)
	ErrInvalidMessageFormat = NewErr(10005, "Invalid message format", http.StatusUnauthorized)
)

var codeToErr = map[uint32]*Err{
	200:   NoErr,
	7000:  ErrCustom,
	7777:  ErrUnexpected,
	9999:  ErrTokenNotValidYet,
	10002: ErrInvalidParams,
	10003: ErrTokenVerify,
	10004: ErrTokenExpire,
}

// NewErr 创建新的业务错误
func NewErr(code uint32, msg string, httpCode ...int) *Err {
	hc := http.StatusOK
	if len(httpCode) != 0 {
		hc = httpCode[0]
	}

	return &Err{code: code, httpCode: hc, msg: msg}
}

func GetCodeToErr() map[uint32]*Err {
	return codeToErr
}

func SetCodeToErr(code uint32, err *Err) error {
	if _, ok := codeToErr[code]; ok {
		return errors.New("has exist")
	}

	codeToErr[code] = err
	return nil
}

// WithMsg 基于已有错误码创建携带动态消息的新实例
func (e *Err) WithMsg(msg string) *Err {
	return &Err{code: e.code, httpCode: e.httpCode, msg: msg}
}

// NewCustomErr 创建新的自定义错误
func NewCustomErr(msg string, httpCode ...int) *Err {
	return NewErr(CodeCustom, msg, httpCode...)
}

// IsErr 判断是否为业务错误
func IsErr(err error) bool {
	if err == nil {
		return true
	}

	_, ok := err.(*Err)
	return ok
}

// ParseErr 解析业务错误
func ParseErr(err error) *Err {
	if err == nil {
		return NoErr
	}

	/*
		 errors.Wrap 包裹后，错误类型变成了 *errors.withStack（pkg/errors 的内部类型），而这里的 ParseErr 函数只做了 err.(*Err) 的直接类型断言，
		无法穿透 wrap 层识别出底层的业务错误。最终走到了 ParseCode 的兜底逻辑，返回了 ErrUnexpected (7777)
	*/
	// 新增：用 errors.As 穿透 wrap 层查找业务错误
	var bizErr *Err
	if errors.As(err, &bizErr) {
		return bizErr
	}

	// 如果是业务类型结构，直接返回
	if e, ok := err.(*Err); ok {
		return e
	}
	// 如果不是继续解析
	s, _ := status.FromError(err)
	c := uint32(s.Code())
	if c == CodeCustom {
		return NewCustomErr(s.Message())
	}
	// 最后检查是不是业务类型 返回码
	return ParseCode(c)
}

// ParseCode 解析业务状态码对应的业务错误
func ParseCode(code uint32) *Err {
	if e, ok := codeToErr[code]; ok {
		return e
	}
	// 如果不是业务类型返回码，直接返回固定错误
	return ErrUnexpected
}
