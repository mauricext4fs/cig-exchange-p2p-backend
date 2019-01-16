package main

import (
	"github.com/joho/godotenv"
	"cig-exchange-p2p-backend/controllers"
	"net/http"

	"os"
	"fmt"
    "strings"
	"github.com/gorilla/mux"
)

func main() {

	e := godotenv.Load()
	if e != nil {
		fmt.Print(e)
	}

	baseUri := os.Getenv("P2P_BACKEND_BASE_URI")
    baseUri = strings.Replace(baseUri, "\"", "", -1)
	fmt.Println("Base URI set to " + baseUri)

    // For some god fucking reason using this does not work in router!!!!
    baseUri = "/p2p/api/"


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
