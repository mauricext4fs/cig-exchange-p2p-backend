package main

import (
	"cig-exchange-libs/auth"
	"cig-exchange-p2p-backend/controllers"
	"net/http"

	"github.com/joho/godotenv"

	"fmt"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

func main() {

	e := godotenv.Load()
	if e != nil {
		fmt.Print(e)
	}

	baseURI := os.Getenv("P2P_BACKEND_BASE_URI")
	baseURI = strings.Replace(baseURI, "\"", "", -1)
	fmt.Println("Base URI set to " + baseURI)

	router := mux.NewRouter()

	// List of endpoints that doesn't require auth
	skipJWT := []string{"ping"}
	userAPI := auth.NewUserAPI(auth.PlatformP2P, baseURI, skipJWT)

	router.HandleFunc(baseURI+"ping", controllers.Ping).Methods("GET")
	router.HandleFunc(baseURI+"offerings", controllers.CreateOffering).Methods("POST")
	router.HandleFunc(baseURI+"offerings", controllers.GetOfferings).Methods("GET")
	router.HandleFunc(baseURI+"offerings/{offering_id}", controllers.GetOffering).Methods("GET")
	router.HandleFunc(baseURI+"offerings/{offering_id}", controllers.UpdateOffering).Methods("PATCH")
	router.HandleFunc(baseURI+"offerings/{offering_id}", controllers.DeleteOffering).Methods("DELETE")

	// attach JWT auth middleware
	router.Use(userAPI.JwtAuthenticationHandler)

	//router.NotFoundHandler = app.NotFoundHandler

	port := os.Getenv("DOCKER_LISTEN_DEFAULT_PORT")
	//port := "80"

	fmt.Println(port)

	err := http.ListenAndServe(":"+port, router) //Launch the app, visit localhost:8000/api
	if err != nil {
		fmt.Print(err)
	}
}
