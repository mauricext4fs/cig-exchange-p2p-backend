package controllers

import (
	u "cig-exchange-p2p-backend/utils"
	"net/http"
)

var Ping = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "Pong!")
	u.Respond(w, resp)
}
