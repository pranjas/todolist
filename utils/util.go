package utils

import "context"

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
