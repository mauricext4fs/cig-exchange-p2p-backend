package controllers

import (
	cigExchange "cig-exchange-libs"
	"cig-exchange-libs/auth"
	"cig-exchange-libs/models"
	"encoding/json"
	"net/http"
)

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
