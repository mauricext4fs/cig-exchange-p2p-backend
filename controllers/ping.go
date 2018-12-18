package controllers

import (
	"net/http"
	u "cig-exchange-sso-backend/utils"
)

var Ping = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "Pong!")
	u.Respond(w, resp)
}
