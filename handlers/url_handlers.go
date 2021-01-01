package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
	"todolist/database"
	"todolist/environment"
	"todolist/handlers/token"
	"todolist/model"
	"todolist/responses"
	"todolist/tptverify"
	"todolist/utils"

	"go.mongodb.org/mongo-driver/mongo"
)

func Login(w http.ResponseWriter, r *http.Request) {
	//Get a Database connection and check for the
	//userID and password provided.

	if redirectToHTTPS(&w, r) ||
		!checkRequestMethod(&w, r, http.MethodPost) ||
		!checkRequestHeaders(&w, r) {
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		GenericInternalServerHeader(&w, r)
		return
	}

	user := &model.User{}
	//Try to fill in the empty user struct from the
	//json recieved.
	err = json.Unmarshal(bytes, user)
	if err != nil {
		GenericBadRequest(&w, "json body contains unidentified members.")
		return
	}
	connection, err := database.GetMongoConnection(environment.GetMongoConnectionString())
	if err != nil {
		log.Printf("Error getting mongo connection")
		GenericInternalServerError(&w, "Internal server error")
		return
	}
	//defer statements execute when the function returns.
	//defer calls are stacked but the arguments are evaluated
	//when a defer statement is encountered and NOT when it's
	//executed. Thus connection here is already evaluated but
	//function call is made when function returns. Changing
	//connection variable post the defer statement WILL NOT
	//cause the function to use the new value.
	defer database.ReleaseMongoConnection(connection)
	realUser := model.GetUser(connection, user.ID, user.Password)
	if realUser == nil {
		GenericBadRequest(&w, "User not found.")
		return
	}
	response := responses.Response{
		Status:  http.StatusOK,
		Message: "Login Successful",
	}
	bearerToken, err := token.GenerateToken(user, user.SignInType)
	if err != nil {
		goto out
	}
	w.Header().Add("Authorization", "Bearer "+bearerToken)
	GenericWriteResponse(&w, &response)
	return
out:
	GenericInternalServerHeader(&w, r)
}

func Register(w http.ResponseWriter, r *http.Request) {

	if redirectToHTTPS(&w, r) ||
		!checkRequestMethod(&w, r, http.MethodPost) ||
		!checkRequestHeaders(&w, r) {
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		GenericInternalServerHeader(&w, r)
		return
	}
	user := &model.User{}
	err = json.Unmarshal(bytes, user)
	if err != nil {
		GenericBadRequest(&w, "json body contains unidentified members.")
		return
	}
	user.SignInType = model.WebLogin
	connection, err := database.GetMongoConnection(environment.GetMongoConnectionString())
	if err != nil {
		GenericInternalServerHeader(&w, r)
		return
	}
	defer database.ReleaseMongoConnection(connection)
	if model.AddUser(connection, user) {
		GenericResponse(&w, "User Registration Successful.", http.StatusOK)
		return
	}
	GenericInternalServerError(&w, "Internal server error.")
}

//TPTVerify endpoint verifies a third party generated token,
//viz google, facebook etc. Since we want to keep the endpoint
//same and short the actual work is done in the type that implements
//the Verifier interface.
func TPTVerify(w http.ResponseWriter, r *http.Request) {
	expected := struct {
		Client       string `json:"client"`
		Os           string `json:"os"`
		Board        string `json:"board"`
		Manufacturer string `json:"manufacturer"`
		Model        string `json:"model"`
	}{}
	if redirectToHTTPS(&w, r) ||
		!checkRequestMethod(&w, r, http.MethodPost) ||
		!checkRequestHeaders(&w, r) {
		log.Print("Missing headers\n")
		return
	}
	bearerToken := token.GetBearerToken(r)

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		GenericInternalServerHeader(&w, r)
		return
	}
	err = json.Unmarshal(bytes, &expected)
	if err != nil {
		GenericBadRequest(&w, "json body contains unidentified members.")
		return
	}
	authProvider, err := utils.GetRequestHeader(r, "X-Resource-Auth")
	if err != nil {
		GenericBadRequest(&w, "Required Authorization header missing.")
		return
	}
	provider, err := tptverify.GetVerifier(authProvider)
	if err != nil {
		GenericResponseWithEC(&w, "Unknown auth provider",
			http.StatusBadRequest, API_ERROR_CODE_INVALID_INPUT)
		log.Printf("Error = %s", err)
		return
	}
	claims, err := provider.Verify(bearerToken)
	if err != nil {
		GenericResponseWithEC(&w, "expired token",
			http.StatusBadRequest, API_ERROR_CODE_TOKEN_EXPIRED)
		log.Printf("Error = %s", err)
		return
	}
	responseMap := provider.ResponseMap(claims)
	response := responses.Response{
		Status:  http.StatusOK,
		Message: fmt.Sprintf("Verified %s Token", provider.Name()),
		Meta:    responseMap,
	}
	connection, err := database.GetMongoConnection(environment.GetMongoConnectionString())
	if err != nil {
		log.Printf("Error getting mongo connection")
		GenericInternalServerError(&w, "Internal server error")
		return
	}
	//defer statements execute when the function returns.
	//defer calls are stacked but the arguments are evaluated
	//when a defer statement is encountered and NOT when it's
	//executed. Thus connection here is already evaluated but
	//function call is made when function returns. Changing
	//connection variable post the defer statement WILL NOT
	//cause the function to use the new value.
	defer database.ReleaseMongoConnection(connection)
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	userid, err := provider.UserId(claims)
	if err != nil {
		log.Printf("Error getting userid, err = %v\n", err)
		GenericInternalServerError(&w, "Internal server error")
		return
	}
	user := model.GetUserForId(connection, userid)
	if user != nil {
		log.Printf("User with id %s already registered. From %s\n",
			userid, provider.Name())
		goto done
	}
	//Attempt to register this new user.
	user = &model.User{}
	user.ID = userid
	user.Meta = responseMap
	user.Password = fmt.Sprintf("%x", rand.Int63())
	user.SignInType, _ = model.GetLoginType(authProvider)
	model.AddUser(connection, user)
done:
	w.Header().Add("Authorization", "Bearer "+bearerToken)
	GenericWriteResponse(&w, &response)
}

func User(w http.ResponseWriter, r *http.Request) {
	GenericWriteHeader(&w, r, http.StatusNotImplemented)
}

func getUserID(w *http.ResponseWriter, r *http.Request, method string) (bool, string) {
	var ok bool
	var userID string
	var err error
	if redirectToHTTPS(w, r) ||
		!checkRequestMethod(w, r, method) ||
		!checkRequestHeaders(w, r) {
		log.Print("Missing headers\n")
		return ok, userID
	}
	claims, response := VerifyBearerToken(r)
	if claims.Claims == nil || claims.Verifier == nil {
		log.Print("Couldn't verify bearer token\n")
		GenericWriteResponse(w, &response)
		return ok, userID
	}
	userID, err = claims.UserId(claims.Claims)
	if err != nil {
		log.Printf("UserID doesn't exist\n")
		GenericBadRequest(w, "User ID doesn't exists")
		return ok, userID
	}
	ok = true
	return ok, userID
}

func postAddOrModify(w *http.ResponseWriter, r *http.Request, modify bool) {
	ok, userID := getUserID(w, r, http.MethodPost)
	if !ok {
		log.Printf("Error extracting userID from request\n")
		return
	}
	connection, err := database.GetMongoConnection(environment.GetMongoConnectionString())
	if err != nil {
		log.Printf("Couldn't get MongoDB Connection\n")
		GenericInternalServerError(w, "An Internal Server Error occured")
		return
	}
	defer database.ReleaseMongoConnection(connection)

	expected := model.TodoItem{}
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		GenericInternalServerHeader(w, r)
		return
	}
	err = json.Unmarshal(bytes, &expected)
	if err != nil {
		GenericBadRequest(w, "json body contains unidentified members.")
		return
	}
	user := model.GetUserForId(connection, userID)
	if user == nil {
		log.Printf("User %s not found\n", userID)
		GenericResponseWithEC(w, "User not found", http.StatusNotFound, API_ERROR_CODE_INVALID_INPUT)
		return
	}
	expected.Owner = userID
	debugText := "add"
	var op func(*model.TodoItem, *mongo.Client) bool
	op = (*model.TodoItem).Add
	if modify {
		debugText = "modify"
		op = (*model.TodoItem).Modify
	}
	if !op(&expected, connection) {
		log.Printf("Couldn't %s ToDo Item for user %s", debugText, userID)
		GenericInternalServerError(w, "A Server Error occured trying to modify / add TodoItem.")
		return
	}
	log.Printf("%s a ToDo Item for user %s\n", debugText, userID)
	GenericResponse(w, fmt.Sprintf("%s Todo Item succeeded", debugText), http.StatusOK)
}

//JSON body contains the Post data.
func PostAdd(w http.ResponseWriter, r *http.Request) {
	postAddOrModify(&w, r, false)
}

//JSON body contains the Post data.
func PostEdit(w http.ResponseWriter, r *http.Request) {
	postAddOrModify(&w, r, true)
}

//JSON Body contains the POST ID
//If the owner ID doesn't match this user's ID
//then the post is attempted to be removed from the
//shared with.
func PostRemove(w http.ResponseWriter, r *http.Request) {
	ok, userID := getUserID(&w, r, http.MethodPost)
	if !ok {
		log.Printf("Error extracting userID from request\n")
		return
	}
	connection, err := database.GetMongoConnection(environment.GetMongoConnectionString())
	if err != nil {
		log.Printf("Couldn't get MongoDB Connection\n")
		GenericInternalServerError(&w, "An Internal Server Error occured")
		return
	}
	defer database.ReleaseMongoConnection(connection)
	expected := struct {
		PostID string `json:"id"`
	}{}
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		GenericInternalServerHeader(&w, r)
		return
	}
	err = json.Unmarshal(bytes, &expected)
	if err != nil {
		GenericBadRequest(&w, "json body contains unidentified members.")
		return
	}
	dummyPostObj := model.TodoItem{
		Owner: userID,
		ID:    expected.PostID,
	}
	removedShared := false
	if !dummyPostObj.Remove(connection) {
		log.Printf("Item %s not owned by user %s", expected.PostID, userID)
		if !dummyPostObj.RemoveFromShared(connection, userID) {
			GenericBadRequest(&w, "Post not shared with user")
			return
		}
		removedShared = true
	}
	log.Printf("Removed ToDo Item %s for user %s, is shared = %s\n",
		expected.PostID, userID, removedShared)
	GenericResponse(&w, "Removed ToDo Item", http.StatusOK)
}

func PostGet(w http.ResponseWriter, r *http.Request) {
	ok, userID := getUserID(&w, r, http.MethodGet)
	if !ok {
		log.Printf("Error extracting userID from request\n")
		return
	}
	connection, err := database.GetMongoConnection(environment.GetMongoConnectionString())
	if err != nil {
		log.Printf("Couldn't get MongoDB Connection\n")
		GenericInternalServerError(&w, "An Internal Server Error occured")
		return
	}
	defer database.ReleaseMongoConnection(connection)
	getShared := false
	var off uint
	var count uint
	if shared, err := utils.GetRequestParam(r, "shared"); err == nil {
		if shared == "1" {
			getShared = true
		}
	}
	if offset, err := utils.GetRequestParam(r, "offset"); err == nil {
		off = utils.ToUint(offset)
	}
	if cnt, err := utils.GetRequestParam(r, "count"); err == nil {
		count = utils.ToUint(cnt)
	}
	//get offset and count in request parameter
	//return count, more and list of items

	items, err := model.GetOwnerItems(connection, userID, getShared, off, count)
	if err != nil {
		GenericInternalServerError(&w, "Unable to process request.")
		return
	}
	resp := responses.Response{
		Status:  http.StatusOK,
		APICode: API_ERROR_CODE_OK,
		Message: "Items fetch complete",
		Meta:    map[string]interface{}{"count": len(items), "items": items},
	}
	GenericWriteResponse(&w, &resp)
}
