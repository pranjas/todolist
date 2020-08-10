package utils

const (
	//Port is the port where we need to listen for requests.
	Port = "PORT"
	//HerokuForwardedProto gives the protocol used by a client
	//to access the resource. We'll use it to redirect client
	//to use https in case http scheme is used.
	HerokuForwardedProto = "X-Forwarded-Proto"
)
