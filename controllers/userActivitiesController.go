package controllers

import (
	cigExchange "cig-exchange-libs"
	"cig-exchange-libs/auth"
	"cig-exchange-libs/models"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// GetUserActivities handles GET users/{user_id}/activities endpoints
var GetUserActivities = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeGetUserActivities)
	defer cigExchange.PrintAPIError(apiErrorP)

	// get request params
	userID := mux.Vars(r)["user_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		*apiErrorP = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	*loggedInUserP = loggedInUser

	// check user id
	if len(userID) == 0 {
		*apiErrorP = cigExchange.NewInvalidFieldError("user_id", "UserID is invalid")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	if userID != loggedInUser.UserUUID {
		*apiErrorP = cigExchange.NewAccessRightsError("No access rights for the user")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	userActs, apiError := models.GetActivitiesForUser(userID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	cigExchange.Respond(w, userActs)
}

// CreateUserActivity handles POST users/activities and users/{user_id}/activities endpoints
var CreateUserActivity = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()

	defer cigExchange.PrintAPIError(apiErrorP)

	// check jwt
	loggedInUser, err := auth.GetContextValues(r)
	if err == nil {
		*loggedInUserP = loggedInUser
	}

	info := make(map[string]interface{})
	// decode organisation object from request body
	err = json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		*apiErrorP = cigExchange.NewJSONDecodingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeCreateUserActivity)
		return
	}

	// insert user activity into db
	apiError := auth.CreateCustomUserActivity(loggedInUserP, info)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	w.WriteHeader(204)
}
