package utils

import (
	"context"
	"net/http"
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
