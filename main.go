package main

import (
	"cig-exchange-libs/auth"
	"cig-exchange-p2p-backend/controllers"
	"cig-exchange-p2p-backend/tasks"
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
	router.HandleFunc(p2pBaseURI+"me/info", userAPI.GetInfo).Methods("GET")
	router.HandleFunc(p2pBaseURI+"ping-jwt", userAPI.PingJWT).Methods("GET")
	router.HandleFunc(p2pBaseURI+"users/switch/{organisation_id}", userAPI.ChangeOrganisationHandler).Methods("POST")
	router.HandleFunc(p2pBaseURI+"users/{user_id}/contacts", controllers.GetContacts).Methods("GET")
	router.HandleFunc(p2pBaseURI+"users/{user_id}/contacts", controllers.CreateContact).Methods("POST")
	router.HandleFunc(p2pBaseURI+"users/{user_id}/contacts/{contact_id}", controllers.UpdateContact).Methods("PATCH")
	router.HandleFunc(p2pBaseURI+"users/{user_id}/contacts/{contact_id}", controllers.DeleteContact).Methods("DELETE")
	router.HandleFunc(p2pBaseURI+"users/{user_id}", controllers.GetUser).Methods("GET")
	router.HandleFunc(p2pBaseURI+"users/{user_id}", controllers.UpdateUser).Methods("PATCH")
	router.HandleFunc(p2pBaseURI+"users/{user_id}/activities", controllers.GetUserActivities).Methods("GET")
	router.HandleFunc(p2pBaseURI+"users/{user_id}/activities", controllers.CreateUserActivity).Methods("POST")
	router.HandleFunc(p2pBaseURI+"organisations", controllers.CreateOrganisation).Methods("POST")                     // only admin can create organisation. Organisation will be empty
	router.HandleFunc(p2pBaseURI+"organisations", controllers.GetOrganisations).Methods("GET")                        // all user will receive list of their organisations, admin will receive all organisations
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}", controllers.GetOrganisation).Methods("GET")       // users can get organisation that they belongs to, admin can get any organisation
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}", controllers.UpdateOrganisation).Methods("PATCH")  // org admins can change all except 'status', admin can change all
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}", controllers.DeleteOrganisation).Methods("DELETE") // admin can delete organisation
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/dashboard", controllers.GetDashboardInfo).Methods("GET")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/dashboard/users", controllers.GetDashboardUsersInfo).Methods("GET")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/dashboard/offerings", controllers.GetDashboardOfferingsBreakdown).Methods("GET")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/dashboard/clicks", controllers.GetDashboardOfferingsClicks).Methods("GET")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/offerings", controllers.CreateOffering).Methods("POST")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/offerings", controllers.GetOfferings).Methods("GET")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/offerings/{offering_id}", controllers.GetOffering).Methods("GET")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/offerings/{offering_id}", controllers.UpdateOffering).Methods("PATCH")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/offerings/{offering_id}", controllers.DeleteOffering).Methods("DELETE")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/offerings/{offering_id}/media/upload", controllers.UploadMedia).Methods("PUT")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/offerings/{offering_id}/media/ordering", controllers.UpdateMediaOrdering).Methods("POST")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/offerings/{offering_id}/media/{media_id}", controllers.UpdateOfferingMedia).Methods("PATCH")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/offerings/{offering_id}/media/{media_id}", controllers.DeleteOfferingMedia).Methods("DELETE")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/users", controllers.GetOrganisationUsers).Methods("GET")                // admin can receive users for any organisation, any user from organisation can see other members
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/users/{user_id}", controllers.DeleteOrganisationUser).Methods("DELETE") // admin can delete any user, org admin can't delete himself
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/users/{user_id}", controllers.AddOrganisationUser).Methods("POST")      // admin can add user to organisation
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/users/{user_id}", controllers.ChangeOrganisationUser).Methods("PATCH")  // organisation admin can add and remove other organisation user
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/invitations", controllers.GetInvitations).Methods("GET")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/invitations", controllers.SendInvitation).Methods("POST")
	router.HandleFunc(p2pBaseURI+"organisations/{organisation_id}/invitations/{user_id}", controllers.DeleteInvitation).Methods("DELETE")

	// trading
	router.HandleFunc(tradingBaseURI+"ping", controllers.Ping).Methods("GET")
	router.HandleFunc(tradingBaseURI+"users/activities", controllers.CreateUserActivity).Methods("POST")
	router.HandleFunc(tradingBaseURI+"users/signup", userAPI.CreateUserHandler).Methods("POST")
	router.HandleFunc(tradingBaseURI+"users/pingdom/signup", userAPI.CreateUserHandlerPingdom).Methods("POST")
	router.HandleFunc(tradingBaseURI+"users/signup/{user_id}/webauthn", userAPI.CreateUserWebAuthnHandler).Methods("POST")
	router.HandleFunc(tradingBaseURI+"users/signin", userAPI.GetUserHandler).Methods("POST")
	router.HandleFunc(tradingBaseURI+"users/signin/{user_id}/webauthn", userAPI.GetUserWebAuthnHandler).Methods("POST")
	router.HandleFunc(tradingBaseURI+"users/send_otp", userAPI.SendCodeHandler).Methods("POST")
	router.HandleFunc(tradingBaseURI+"users/verify_otp", userAPI.VerifyCodeHandler).Methods("POST")
	router.HandleFunc(tradingBaseURI+"users/accept-invitation", controllers.AcceptInvitation).Methods("POST")
	router.HandleFunc(tradingBaseURI+"organisations/signup", userAPI.CreateOrganisationHandler).Methods("POST")
	router.HandleFunc(tradingBaseURI+"offerings", controllers.GetAllOfferings).Methods("GET")
	router.HandleFunc(tradingBaseURI+"offerings/{offering_id}/media", controllers.GetOfferingMedia).Methods("GET")
	router.HandleFunc(tradingBaseURI+"media/{media_file}", controllers.GetMedia).Methods("GET")
	router.HandleFunc(tradingBaseURI+"contact_us", controllers.SendContactUsEmail).Methods("POST")

	//router.NotFoundHandler = app.NotFoundHandler

	port := os.Getenv("DOCKER_LISTEN_DEFAULT_PORT")

	fmt.Println("Server listening on port: " + port)

	// shedule tasks
	tasks.ScheduleTasks()

	// launch the app
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		fmt.Print(err)
	}
}
