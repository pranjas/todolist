package model

import (
	"log"
	"todolist/database"
	"todolist/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type LoginType int

const (
	WebLogin = LoginType(iota)
	GoogleLogin
	FacebookLogin
	TwitterLogin
	GithubLogin
	GitlabLogin
)

type User struct {
	//This is a unique field.
	ID         string                 `json:"id"`
	Meta       map[string]interface{} `json:"extra",omitempty  bson:"extra",omitempty`
	SignInType LoginType              `json:"-" bson:"type"`
	Password   string                 `json:"pass"`
}

//See if we can get a login Id and password
//to match anything in the database.
func GetUser(dbClient *mongo.Client, id, password string) *User {
	u := &User{}
	//Create a mongo query to find a user with
	//matching id and password.
	//For complex queries we'll use bson.D but since
	//this is simple we use bson.M (map)
	context := utils.GetContext()
	query := bson.M{
		"id":       id,
		"password": password,
	}
	collection := database.GetUserCollection(dbClient)
	err := collection.FindOne(context, query).Decode(u)
	if err == nil {
		return u
	}
	return nil
}

func AddUser(dbClient *mongo.Client, user *User) bool {
	context := utils.GetContext()
	if GetUser(dbClient, user.ID, user.Password) != nil {
		return false
	}
	collection := database.GetUserCollection(dbClient)
	res, err := collection.InsertOne(context, *user)
	if err != nil {
		log.Printf("Error adding user %v: %s", *user, err)
		return false
	}
	log.Printf("Added user %v to users collection with object id = %v\n", *user, res.InsertedID)
	return true
}

func (u *User) Update(dbClient *mongo.Client) {
	context := utils.GetContext()
	query := bson.M{
		"id":       u.ID,
		"password": u.Password,
	}
	collection := database.GetUserCollection(dbClient)
	result, err := collection.ReplaceOne(context, query, *u)
	if err == nil {
		log.Printf("Updated %d user(s)\n", result.MatchedCount)
	}
}
