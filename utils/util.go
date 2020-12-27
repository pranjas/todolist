package utils

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type RawResource interface {
}
type Resource struct {
	sync.Mutex
	RawResource
}

const (
	//HerokuForwardedProto gives the protocol used by a client
	//to access the resource. We'll use it to redirect client
	//to use https in case http scheme is used.
	HerokuForwardedProto = "X-Forwarded-Proto"
)

func GetContext() context.Context {
	return context.Background()
}

type StringSlice []string

func (slice StringSlice) Contains(s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func GetRequestHeader(r *http.Request, header string) (string, error) {
	requestHeaderValues, ok := r.Header[header]
	if !ok {
		return "", errors.Errorf("header %s not found", header)
	}
	return requestHeaderValues[0], nil
}
func GetRequestHeaderSubField(r *http.Request, header, subfield string) (string, error) {
	requestHeaderValues, ok := r.Header[header]
	if !ok {
		return "", errors.Errorf("header %s not found", header)
	}
	for _, value := range requestHeaderValues {
		if strings.Contains(value, subfield) {
			return value, nil
		}
	}
	return "", errors.New("Not Found")
}

func GetRequestParam(r *http.Request, param string) (string, error) {
	keys, ok := r.URL.Query()[param]
	if !ok || len(keys) == 0 {
		return "", errors.Errorf("Parameter %s not found", param)
	}
	return keys[0], nil
}

func ToIntXX(str string, width int) int64 {
	val, err := strconv.ParseInt(str, 10, width)
	if err != nil {
		val = 0
	}
	return val
}

func ToUIntXX(str string, width int) uint64 {
	val, err := strconv.ParseUint(str, 10, width)
	if err != nil {
		val = 0
	}
	return val
}

func ToUInt32(str string) uint32 {
	var val uint32
	val = uint32(ToUIntXX(str, 32))
	return val
}

func ToInt32(str string) int32 {
	var val int32
	val = int32(ToIntXX(str, 32))
	return val
}

func ToUint(str string) uint {
	var val uint
	val = uint(ToIntXX(str, 64))
	return val
}

func ToInt(str string) int {
	var val int

	val = int(ToIntXX(str, 64))
	return val
}
