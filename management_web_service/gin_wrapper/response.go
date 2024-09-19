package gin_wrapper

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

type GinResponse struct {
	wrapper    GinWrapper
	success    bool
	headers    map[string]string
	statusCode int
	result     any
}

func (w GinWrapper) PrepareDefaultSuccessResponse(result any) *GinResponse {
	return &GinResponse{
		wrapper:    w,
		success:    true,
		statusCode: http.StatusOK,
		result:     result,
	}
}

func (w GinWrapper) PrepareDefaultErrorResponse() *GinResponse {
	return &GinResponse{
		wrapper:    w,
		success:    false,
		statusCode: http.StatusInternalServerError,
		result:     "internal server error",
	}
}

func (w GinWrapper) PrepareResponseBadBinding(err error) *GinResponse {
	return w.PrepareDefaultErrorResponse().
		WithHttpStatusCode(http.StatusBadRequest).
		WithResult(errors.Wrap(err, "unable to bind your request").Error())
}

func (r *GinResponse) WithHttpStatusCode(statusCode int) *GinResponse {
	r.statusCode = statusCode
	return r
}

func (r *GinResponse) WithHeader(key, value string) *GinResponse {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[key] = value
	return r
}

func (r *GinResponse) WithResult(result string) *GinResponse {
	r.result = result
	return r
}

func (r GinResponse) SendResponse() {
	if len(r.headers) > 0 {
		for k, v := range r.headers {
			r.wrapper.c.Header(k, v)
		}
	}

	var status, message string
	if r.success {
		status = "1"
		message = "OK"
	} else {
		status = "0"
		//goland:noinspection SpellCheckingInspection
		message = "NOTOK"
	}
	r.wrapper.c.JSON(r.statusCode, gin.H{
		"status":  status,
		"message": message,
		"result":  r.result,
	})
}
