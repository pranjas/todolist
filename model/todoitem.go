package model

import (
	"log"
	"todolist/database"
	"todolist/utils"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TodoItem struct {
	Owner     string                 `json:"-"`
	Name      string                 `json:"name"`
	Content   map[string]interface{} `json:"content,omitempty"`
	Actions   map[string]interface{} `json:"actions",omitempty`
	StartTime string                 `json:"start_time",omitempty`
	EndTime   string                 `json:"end_time", omitempty`
	ID        string                 `json:"id", omitempty`
	//SharedWith contains the userIDs of the users
	//This TodoItem is shared with.
	SharedWith []string `json:"sharedWith", omitempty`
}

var globalLock utils.Resource

func (todoItem *TodoItem) RemoveFromShared(dbClient *mongo.Client, sharedUserID string) bool {
	storedItem := &TodoItem{}
	context := utils.GetContext()
	query := bson.D{
		{"sharedwith", bson.D{{"$in", bson.A{sharedUserID}}}},
		{"ID", todoItem.ID},
	}
	collection := database.GetTodoListCollection(dbClient)
	globalLock.Lock()
	defer globalLock.Unlock()
	err := collection.FindOne(context, query).Decode(storedItem)
	if err != nil {
		log.Printf("TodoItem with ID = %s, not shared with %s\n", todoItem.ID, sharedUserID)
		return false
	}

	log.Printf("Found TodoItem with ID %s, shared with %v\n", storedItem.ID, storedItem.SharedWith)
	var idx int
	//Find where this ID is present in the Index
	for __idx, value := range storedItem.SharedWith {
		if value == sharedUserID {
			idx = __idx
			break
		}
	}
	newSharedUsers := make([]string, len(storedItem.SharedWith)-1)
	newSharedUsers = append(newSharedUsers, storedItem.SharedWith[:idx]...)
	newSharedUsers = append(newSharedUsers, storedItem.SharedWith[idx+1:]...)
	todoItem.SharedWith = newSharedUsers

	return todoItem.Modify(dbClient)
}

func (todoItem *TodoItem) Remove(dbClient *mongo.Client) bool {
	context := utils.GetContext()
	query := bson.M{
		"owner": todoItem.Owner,
		"ID":    todoItem.ID,
	}
	collection := database.GetTodoListCollection(dbClient)
	res, err := collection.DeleteOne(context, query)
	if err != nil {
		log.Printf("No document found for owner %s, with ID = %s\n", todoItem.Owner, todoItem.ID)
		return false
	}
	log.Printf("Removed %d item(s) for owner %s", res.DeletedCount, todoItem.Owner)
	return true
}

func RemoveAllItemsForOwner(dbClient *mongo.Client, owner string) bool {
	context := utils.GetContext()
	query := bson.M{
		"owner": owner,
	}
	collection := database.GetTodoListCollection(dbClient)
	res, err := collection.DeleteMany(context, query)
	if err != nil {
		log.Printf("No document found for owner %s\n", owner)
		return false
	}
	log.Printf("Removed %d item(s) for owner %s", res.DeletedCount, owner)
	return true
}

func (todoItem *TodoItem) Add(dbClient *mongo.Client) bool {
	context := utils.GetContext()
	collection := database.GetTodoListCollection(dbClient)
	res, err := collection.InsertOne(context, todoItem)
	if err != nil {
		log.Printf("Error adding todoItem %v", *todoItem)
		return false
	}
	log.Printf("Added a todoItem, %v with objectId %v\n", *todoItem, res.InsertedID)
	return true
}

func (todoItem *TodoItem) Modify(dbClient *mongo.Client) bool {
	context := utils.GetContext()
	query := bson.M{
		"owner": todoItem.Owner,
		"id":    todoItem.ID,
	}
	collection := database.GetTodoListCollection(dbClient)
	res, err := collection.ReplaceOne(context, query, *todoItem)
	if err != nil {
		log.Printf("Error updating todoItem for owner %s with ID = %s\n", todoItem.Owner, todoItem.ID)
		return false
	}
	log.Printf("Updated %d TodoItem for Owner %s with ID = %s\n", res.MatchedCount, todoItem.Owner, todoItem.ID)
	return true
}

func GetOneTodoItemForOwner(dbClient *mongo.Client, owner, todoItemID string) (*TodoItem, error) {
	query := bson.M{
		"owner": owner,
		"id":    todoItemID,
	}
	context := utils.GetContext()
	collection := database.GetTodoListCollection(dbClient)
	item := &TodoItem{}
	err := collection.FindOne(context, query).Decode(item)
	if err != nil {
		log.Printf("No Item found for user %s with ID %s", owner, todoItemID)
		return nil, errors.Errorf("No Item found for user %s, item ID = %s", owner, todoItemID)
	}
	return item, nil
}

func GetOwnerItems(dbClient *mongo.Client, owner string, getShared bool, off uint, count uint) ([]TodoItem, error) {
	query := bson.M{
		"owner": owner,
	}
	if getShared {
		query["sharedwith"] = bson.D{{"$in", bson.A{owner}}}
	}
	//Create empty Find option
	findOpts := options.Find()

	//If we need to skip things add the relevant
	//option.
	if off > 0 {
		findOpts.SetSkip(int64(off))
	}
	//Limit the total records
	if count > 0 {
		findOpts.SetLimit(int64(count))
	}
	context := utils.GetContext()
	collection := database.GetTodoListCollection(dbClient)
	cursor, err := collection.Find(context, query, findOpts)
	if err != nil {
		log.Printf("No TodoItem found for owner %s", owner)
		return nil, errors.Errorf("No TODO items found for owner %s", owner)
	}
	//We found something let's get it out.
	var todoItems []TodoItem
	err = cursor.All(context, &todoItems)
	defer cursor.Close(context)
	if err != nil {
		log.Printf("Error iterating cursor: %v", err)
		return nil, errors.Errorf("Couldn't decode TodoItems for owner %s", owner)
	}
	return todoItems, nil
}
