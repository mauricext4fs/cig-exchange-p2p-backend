package main

import (
	"cig-exchange-p2p-backend/controllers"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	baseUri := "/invest/api/"

	router := mux.NewRouter()

	router.HandleFunc(baseUri+"ping", controllers.Ping).Methods("GET")
	router.HandleFunc(baseUri+"contacts/new", controllers.CreateContact).Methods("POST")
	router.HandleFunc(baseUri+"me/contacts", controllers.GetContactsFor).Methods("GET") //  user/2/contacts
	router.HandleFunc(baseUri+"offerings", controllers.CreateOffering).Methods("POST")
	router.HandleFunc(baseUri+"offerings", controllers.GetOfferings).Methods("GET")
	router.HandleFunc(baseUri+"offerings/{offering_id}", controllers.GetOffering).Methods("GET")
	router.HandleFunc(baseUri+"offerings/{offering_id}", controllers.UpdateOffering).Methods("PATCH")
	router.HandleFunc(baseUri+"offerings/{offering_id}", controllers.DeleteOffering).Methods("DELETE")

	//attach JWT auth middleware
	//router.Use(app.JwtAuthentication)

	//router.NotFoundHandler = app.NotFoundHandler

	// We always run in docker... for sack of convenience let's always use port 80
	port := "80"

	fmt.Println(port)

	err := http.ListenAndServe(":"+port, router) //Launch the app, visit localhost:8000/api
	if err != nil {
		fmt.Print(err)
	}
}
