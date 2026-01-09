package errors

import (
	"fmt"
	"net/http"
)

// ErrorType 错误类型
type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "validation"   // 验证错误
	ErrorTypeNotFound     ErrorType = "not_found"    // 资源未找到
	ErrorTypeUnauthorized ErrorType = "unauthorized" // 未授权
	ErrorTypeForbidden    ErrorType = "forbidden"    // 禁止访问
	ErrorTypeInternal     ErrorType = "internal"     // 内部错误
	ErrorTypeExternal     ErrorType = "external"     // 外部服务错误
)

// DomainError 领域错误
type DomainError struct {
	Type    ErrorType
	Code    int
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// HTTPStatus 返回对应的HTTP状态码
func (e *DomainError) HTTPStatus() int {
	if e.Code != 0 {
		return e.Code
	}
	switch e.Type {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case ErrorTypeForbidden:
		return http.StatusForbidden
	case ErrorTypeInternal:
		return http.StatusInternalServerError
	case ErrorTypeExternal:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

// NewValidationError 创建验证错误
func NewValidationError(message string, err error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeValidation,
		Code:    http.StatusBadRequest,
		Message: message,
		Err:     err,
	}
}

// NewNotFoundError 创建未找到错误
func NewNotFoundError(resource string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeNotFound,
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

// NewUnauthorizedError 创建未授权错误
func NewUnauthorizedError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeUnauthorized,
		Code:    http.StatusUnauthorized,
		Message: message,
	}
}

// NewForbiddenError 创建禁止访问错误
func NewForbiddenError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeForbidden,
		Code:    http.StatusForbidden,
		Message: message,
	}
}

// NewInternalError 创建内部错误
func NewInternalError(message string, err error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeInternal,
		Code:    http.StatusInternalServerError,
		Message: message,
		Err:     err,
	}
}

// NewExternalError 创建外部服务错误
func NewExternalError(service string, err error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeExternal,
		Code:    http.StatusBadGateway,
		Message: fmt.Sprintf("external service error: %s", service),
		Err:     err,
	}
}

// WrapError 包装错误
func WrapError(err error, message string) *DomainError {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr
	}
	return &DomainError{
		Type:    ErrorTypeInternal,
		Code:    http.StatusInternalServerError,
		Message: message,
		Err:     err,
	}
}
