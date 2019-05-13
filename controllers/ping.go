package controllers

import (
	u "cig-exchange-libs"
	"net/http"
)

// Ping handles the ping test request
var Ping = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "Pong!")
	u.Respond(w, resp)
}
