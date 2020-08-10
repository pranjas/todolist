package handlers

import (
	"net/http"
)

func Login(w http.ResponseWriter, r *http.Request) {
	GenericNotImplemented(w, r)
}

func Register(w http.ResponseWriter, r *http.Request) {
	GenericNotImplemented(w, r)
}

func User(w http.ResponseWriter, r *http.Request) {
	GenericNotImplemented(w, r)
}

func PostAdd(w http.ResponseWriter, r *http.Request) {
	GenericNotImplemented(w, r)
}

func PostRemove(w http.ResponseWriter, r *http.Request) {
	GenericNotImplemented(w, r)
}

func PostEdit(w http.ResponseWriter, r *http.Request) {
	GenericNotImplemented(w, r)
}

func GenericNotImplemented(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
