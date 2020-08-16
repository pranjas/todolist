package main

import (
	"net/http"
	"todolist/environment"
	"todolist/handlers"
)

func main() {
	port := environment.GetPort()
	http.HandleFunc("/login", handlers.Login)
	http.HandleFunc("/register", handlers.Register)
	http.HandleFunc("/user", handlers.User)
	http.HandleFunc("/post/add", handlers.PostAdd)
	http.HandleFunc("/post/remove", handlers.PostRemove)
	http.HandleFunc("/post/edit", handlers.PostEdit)
	http.HandleFunc("/", handlers.GenericNotImplemented)
	http.ListenAndServe(":"+port, nil)
}
