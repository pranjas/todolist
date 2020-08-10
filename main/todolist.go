package main

import (
	"net/http"
	"os"
	"todolist/handlers"
	"todolist/utils"
)

func main() {
	port := os.Getenv(utils.Port)
	http.HandleFunc("/login", handlers.Login)
	http.HandleFunc("/register", handlers.Register)
	http.HandleFunc("/user", handlers.User)
	http.HandleFunc("/post/add", handlers.PostAdd)
	http.HandleFunc("/post/remove", handlers.PostRemove)
	http.HandleFunc("/post/edit", handlers.PostEdit)
	http.HandleFunc("/", handlers.GenericNotImplemented)
	http.ListenAndServe(":"+port, nil)
}
