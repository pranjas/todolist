package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"todolist/database"
	"todolist/model"
	"todolist/utils"
)

func Login(w http.ResponseWriter, r *http.Request) {
	//Get a Database connection and check for the
	//userID and password provided.
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
		GenericWriteHeader(&w, r, http.StatusBadRequest)
		return
	}
	connection, err := database.GetMongoConnection(os.Getenv(utils.MongoDBConnectionString))
	if err != nil {
		GenericInternalServerHeader(&w, r)
		return
	}
	//defer statements execute when the function returns.
	defer database.ReleaseMongoConnection(connection)
	realUser := model.GetUser(connection, user.ID, user.Password)
	if realUser == nil {
		GenericWriteHeader(&w, r, http.StatusBadRequest)
		return
	}
	GenericWriteHeader(&w, r, http.StatusOK)
}

func Register(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		GenericInternalServerHeader(&w, r)
		return
	}
	user := &model.User{}
	err = json.Unmarshal(bytes, user)
	if err != nil {
		GenericWriteHeader(&w, r, http.StatusBadRequest)
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
		GenericWriteHeader(&w, r, http.StatusOK)
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

func GenericInternalServerHeader(w *http.ResponseWriter, r *http.Request) {
	GenericWriteHeader(w, r, http.StatusInternalServerError)
}
func GenericWriteHeader(w *http.ResponseWriter, r *http.Request, code int) {
	(*w).WriteHeader(code)
}
