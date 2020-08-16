package utils

import "context"

const (
	//Port is the port where we need to listen for requests.
	Port = "PORT"
	//HerokuForwardedProto gives the protocol used by a client
	//to access the resource. We'll use it to redirect client
	//to use https in case http scheme is used.
	HerokuForwardedProto    = "X-Forwarded-Proto"
	MongoDBConnectionString = "MONGO_DB_CONNECTION_STRING"
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
