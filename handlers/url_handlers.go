package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"todolist/database"
	"todolist/model"
	"todolist/responses"
	"todolist/utils"
)

//Map the required headers based on whether they are mandatory or not.
//Most of the time a header would contain only a single value but we
//make it an array here to allow for multiple values if any.
type requestHeadersRequired struct {
	Name     string
	Values   []string
	Required bool
}

var headersRequired = map[string][]requestHeadersRequired{
	http.MethodGet: {
		{Name: "Content-Type", Values: []string{"application/json"}, Required: true},
	},
	http.MethodPost: {
		{Name: "Content-Type", Values: []string{"application/json"}, Required: true},
		{Name: "Authorization", Values: []string{"Bearer"}, Required: true},
		{Name: "X_Resource_Auth", Values: []string{""}, Required: false},
	},
}

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
		response := responses.Response{
			Status:  http.StatusBadRequest,
			Message: "json body contains unidentified members.",
		}
		GenericWriteResponse(&w, &response)
		return
	}
	connection, err := database.GetMongoConnection(os.Getenv(utils.MongoDBConnectionString))
	if err != nil {
		GenericInternalServerHeader(&w, r)
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
		response := responses.Response{
			Status:  http.StatusBadRequest,
			Message: "User not found.",
		}
		GenericWriteResponse(&w, &response)
		return
	}
	response := responses.Response{
		Status:  http.StatusOK,
		Message: "Login Successful",
	}
	GenericWriteResponse(&w, &response)
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
		response := responses.Response{
			Status:  http.StatusBadRequest,
			Message: "json body contains unidentified members.",
		}
		GenericWriteResponse(&w, &response)
		return
	}
	user.SignInType = model.WebLogin
	connection, err := database.GetMongoConnection(os.Getenv(utils.MongoDBConnectionString))
	if err != nil {
		GenericInternalServerHeader(&w, r)
		return
	}
	defer database.ReleaseMongoConnection(connection)
	if model.AddUser(connection, user) {
		response := responses.Response{
			Status:  http.StatusOK,
			Message: "User Registration Successful.",
		}
		GenericWriteResponse(&w, &response)
		return
	}
	GenericInternalServerHeader(&w, r)
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

func GenericNotImplemented(w http.ResponseWriter, r *http.Request) {
	GenericWriteHeader(&w, r, http.StatusNotImplemented)
}

func GenericWriteResponse(w *http.ResponseWriter, resp *responses.Response) {
	respBytes, err := json.Marshal(*resp)
	if err != nil {
		(*w).WriteHeader(http.StatusInternalServerError)
		log.Printf("Error Marshalling response %s", resp)
		return
	}
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(resp.Status)
	(*w).Write(respBytes)
}

func GenericInternalServerHeader(w *http.ResponseWriter, r *http.Request) {
	GenericWriteHeader(w, r, http.StatusInternalServerError)
}
func GenericWriteHeader(w *http.ResponseWriter, r *http.Request, code int) {
	(*w).WriteHeader(code)
}

//Heroku gives a header named X-Forwarded-Proto
//which contains the scheme the request originally
//landed on heroku server. NOTE that we don't run
//a HTTPS server, all requests come to us as plain
//HTTP request since it's forwarded internally by
//Heroku to us.
func redirectToHTTPS(w *http.ResponseWriter, r *http.Request) bool {
	scheme, ok := r.Header[utils.HerokuForwardedProto]
	//We're not running behind Heroku or a
	//Cloud based host that supports X-Forwarded-Proto.
	if !ok {
		return false
	}
	//The magic http code is 307 which causes http clients
	//to re-issue request with the correct http method.
	//Not using 307 causes some clients to change the original
	//http method to POST by default.
	if scheme[0] != "https" {
		httpsURL := fmt.Sprintf("https://%s%s", r.Host, r.URL.Path)
		if len(r.URL.RawQuery) > 0 {
			httpsURL = fmt.Sprintf("%s?%s", httpsURL, r.URL.RawQuery)
		}
		http.Redirect(*w, r, httpsURL, http.StatusTemporaryRedirect)
		return true
	}
	return false
}

func checkRequestHeaders(w *http.ResponseWriter, r *http.Request) bool {
	expectedHeaders, ok := headersRequired[r.Method]
	result := true
	if !ok {
		result = false
		goto out
	}
	//For a particular request, go over all expected headers.
	//and match the values.
	for _, header := range expectedHeaders {
		requestHeaderValues, ok := r.Header[header.Name]
		log.Printf("[%s] = %v", header.Name, requestHeaderValues)
		//header not found but is required.
		if !ok && header.Required {
			result = false
			goto out
		}
		//Support only the first value
		//check of request header
		if (requestHeaderValues[0] != header.Values[0]) && header.Required {
			result = false
			goto out
		}
	}
out:
	if !result {
		response := responses.Response{
			Status:  http.StatusBadRequest,
			Message: "Required Request headers missing.",
		}
		GenericWriteResponse(w, &response)
	}
	return result
}

func checkRequestMethod(w *http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		response := responses.Response{
			Status:  http.StatusBadRequest,
			Message: "HTTP Method Not Supported",
		}
		GenericWriteResponse(w, &response)
		return false
	}
	return true
}
