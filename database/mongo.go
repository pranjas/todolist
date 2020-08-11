package database

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//We don't want an unlimited connections to be made to
//mongodb server. Keep a tab on the max number of connections
const (
	maxConnections     = 40
	userDatabase       = "userdb"
	userCollection     = "users"
	todolistDatabase   = "todolistdb"
	todolistCollection = "todolist"
)

//Mutex helps us to keep the conn count synced properly.
var connMutex sync.Mutex
var connCount int

func GetUserCollection(dbClient *mongo.Client) *mongo.Collection {
	collection := dbClient.Database(userDatabase).Collection(userCollection)
	return collection
}

func GetTodoListCollection(dbClient *mongo.Client) *mongo.Collection {
	collection := dbClient.Database(todolistDatabase).Collection(todolistCollection)
	return collection
}

func ReleaseMongoConnection(client *mongo.Client) {
	log.Printf("Releasing db connection\n")
	if client != nil {
		client.Disconnect(context.Background())
	}

	connMutex.Lock()
	defer connMutex.Unlock()
	connCount -= 1
	log.Printf("Released db connection\n")
}

func GetMongoConnection(mongoURI string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	connMutex.Lock()
	defer connMutex.Unlock()
	if connCount == maxConnections {
		return nil, errors.New("Max DB Connection limit reached")
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		mongoURI))
	if err != nil {
		log.Printf("Error connection to %s, err = %v\n", mongoURI, err)
	} else {
		connCount += 1
		log.Printf("Connection was successful to mongodb, current connections = %d",
			connCount)
	}
	return client, err
}
