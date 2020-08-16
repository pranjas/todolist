package environment

import "os"

const (
	//Port is the port where we need to listen for requests.
	Port                    = "PORT"
	MongoDBConnectionString = "MONGO_DB_CONNECTION_STRING"
	AppTokenSecret          = "APP_TOKEN_SECRET"
	AppName                 = "APP_NAME"
)

func GetEnvironment(variable string) string {
	return os.Getenv(variable)
}

func GetPort() string {
	return GetEnvironment(Port)
}
func GetAppTokenSecret() string {
	return GetEnvironment(AppTokenSecret)
}
func GetMongoConnectionString() string {
	return GetEnvironment(MongoDBConnectionString)
}

func GetAppName() string {
	appName := GetEnvironment(AppName)
	if appName == "" {
		appName = "this_app"
	}
	return appName
}
