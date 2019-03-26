package main

import (
	"cig-exchange-libs/auth"
	"cig-exchange-p2p-backend/controllers"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {

	e := godotenv.Load()
	if e != nil {
		fmt.Print(e)
	}

	p2pBaseURI := os.Getenv("P2P_BACKEND_BASE_URI")
	p2pBaseURI = strings.Replace(p2pBaseURI, "\"", "", -1)
	fmt.Println("p2p base URI set to " + p2pBaseURI)

	tradingBaseURI := os.Getenv("HOMEPAGE_BACKEND_BASE_URI")
	tradingBaseURI = strings.Replace(tradingBaseURI, "\"", "", -1)
	fmt.Println("trading base URI set to " + tradingBaseURI)

	router := mux.NewRouter()

	userAPI := auth.UserAPI{
		SkipPrefix: tradingBaseURI,
	}

	// register handlers for both platforms
	// p2p
	router.HandleFunc(p2pBaseURI+"users/switch/{organisation_id}", userAPI.ChangeOrganisationHandler).Methods("POST")
	router.HandleFunc(p2pBaseURI+"organisations", controllers.CreateOrganisation).Methods("POST")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}", controllers.GetOrganisation).Methods("GET")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}", controllers.UpdateOrganisation).Methods("PATCH")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}", controllers.DeleteOrganisation).Methods("DELETE")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/offerings", controllers.CreateOffering).Methods("POST")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/offerings", controllers.GetOfferings).Methods("GET")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/offerings/{offering_id}", controllers.GetOffering).Methods("GET")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/offerings/{offering_id}", controllers.UpdateOffering).Methods("PATCH")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/offerings/{offering_id}", controllers.DeleteOffering).Methods("DELETE")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/users", controllers.GetOrganisationUsers).Methods("GET")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/users/{user_id}", controllers.DeleteOrganisationUser).Methods("DELETE")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/invitations", controllers.GetInvitations).Methods("GET")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/invitations", controllers.SendInvitation).Methods("POST")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/invitations/{user_id}", controllers.DeleteInvitation).Methods("DELETE")

	// trading
	router.HandleFunc(tradingBaseURI+"ping", controllers.Ping).Methods("GET")
	router.HandleFunc(tradingBaseURI+"users/signup", userAPI.CreateUserHandler).Methods("POST")
	router.HandleFunc(tradingBaseURI+"users/signin", userAPI.GetUserHandler).Methods("POST")
	router.HandleFunc(tradingBaseURI+"users/send_otp", userAPI.SendCodeHandler).Methods("POST")
	router.HandleFunc(tradingBaseURI+"users/verify_otp", userAPI.VerifyCodeHandler).Methods("POST")
	router.HandleFunc(tradingBaseURI+"organisations/signup", userAPI.CreateOrganisationHandler).Methods("POST")
	router.HandleFunc(tradingBaseURI+"offerings", controllers.GetAllOfferings).Methods("GET")
	router.HandleFunc(tradingBaseURI+"contact_us", controllers.SendContactUsEmail).Methods("POST")

	// attach JWT auth middleware
	router.Use(userAPI.JwtAuthenticationHandler)

	//router.NotFoundHandler = app.NotFoundHandler

	port := os.Getenv("DOCKER_LISTEN_DEFAULT_PORT")

	fmt.Println("Server listening on port: " + port)

	err := http.ListenAndServe(":"+port, router) //Launch the app, visit localhost:8000/api
	if err != nil {
		fmt.Print(err)
	}
}
