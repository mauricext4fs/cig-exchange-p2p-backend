package main

import (
	"github.com/gorilla/mux"
        //"cig-exchange-sso-backend/app"
	"os"
	"fmt"
	"net/http"
	"cig-exchange-sso-backend/controllers"
)

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/api/ping", controllers.Ping).Methods("GET")
	router.HandleFunc("/api/user/new", controllers.CreateAccount).Methods("POST")
	router.HandleFunc("/api/user/login", controllers.Authenticate).Methods("POST")
	router.HandleFunc("/api/contacts/new", controllers.CreateContact).Methods("POST")
	router.HandleFunc("/api/me/contacts", controllers.GetContactsFor).Methods("GET") //  user/2/contacts

        //attach JWT auth middleware
        router.Use(app.JwtAuthentication)

	//router.NotFoundHandler = app.NotFoundHandler

	port := os.Getenv("SSO_BACKEND_PORT")
	if port == "" {
		port = "8000" //localhost
	}

	fmt.Println(port)

	err := http.ListenAndServe(":" + port, router) //Launch the app, visit localhost:8000/api
	if err != nil {
		fmt.Print(err)
	}
}
