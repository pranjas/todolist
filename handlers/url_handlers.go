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
		Message: "Verified Google Token",
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
		log.Printf("Google user with id %s already registered\n",
			userid)
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

func PostAdd(w http.ResponseWriter, r *http.Request) {
	GenericWriteHeader(&w, r, http.StatusNotImplemented)
}

func PostRemove(w http.ResponseWriter, r *http.Request) {
	GenericWriteHeader(&w, r, http.StatusNotImplemented)
}

func PostEdit(w http.ResponseWriter, r *http.Request) {
	GenericWriteHeader(&w, r, http.StatusNotImplemented)
}
