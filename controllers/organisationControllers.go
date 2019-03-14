package controllers

import (
	"cig-exchange-libs"
	"cig-exchange-libs/auth"
	"cig-exchange-libs/models"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// GetOrganisationUsers handles GET organisations/{organisation_id}/users endpoint
var GetOrganisationUsers = func(w http.ResponseWriter, r *http.Request) {

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		apiError := cigExchange.NewRoutingError(err)
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// check permissions
	if organisationID != loggedInUser.OrganisationUUID {
		apiError := cigExchange.NewAccessRightsError("No access rights for the organisation")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// query users from db
	users, apiError := models.GetUsersForOrganisation(organisationID)
	if err != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	cigExchange.Respond(w, users)
}

// DeleteOrganisationUser handles DELETE organisations/{organisation_id}/users/{user_id} endpoint
var DeleteOrganisationUser = func(w http.ResponseWriter, r *http.Request) {

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	userID := mux.Vars(r)["user_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		apiError := cigExchange.NewRoutingError(err)
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// check permissions
	if organisationID != loggedInUser.OrganisationUUID {
		apiError := cigExchange.NewAccessRightsError("No access rights for the organisation")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// find user
	_, apiError := models.GetUser(userID)
	if err != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}
	// TODO: rewoke user JWT token

	// delete OrganizationUser
	orgUser := models.OrganisationUser{
		UserID:         userID,
		OrganisationID: organisationID,
	}

	// delete OrganisationUser
	apiError = orgUser.Delete()
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}
	w.WriteHeader(204)
}
