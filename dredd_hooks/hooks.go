package main

import (
	"encoding/json"

	"github.com/snikch/goodman/hooks"
	trans "github.com/snikch/goodman/transaction"
)

func main() {

	h := hooks.NewHooks()
	server := hooks.NewServer(hooks.NewHooksRunner(h))

	// save created record UUID here
	createdUUID := ""

	h.After("Offerings > api/offerings > Create offering", func(t *trans.Transaction) {
		// happens when api is down
		if t.Real == nil {
			return
		}

		body := map[string]interface{}{}

		json.Unmarshal([]byte(t.Real.Body), &body)
		createdUUID = body["id"].(string)

	})

	// update URI everywhere to point to a created record
	h.Before("Offerings > api/offerings/{offering_id} > Retrieve offering", func(t *trans.Transaction) {
		t.Request.URI = "/api/offerings/" + createdUUID
		t.FullPath = "/api/offerings/" + createdUUID
	})

	h.Before("Offerings > api/offerings/{offering_id} > Update offering", func(t *trans.Transaction) {
		t.Request.URI = "/api/offerings/" + createdUUID
		t.FullPath = "/api/offerings/" + createdUUID
	})

	h.Before("Offerings > api/offerings/{offering_id} > Delete offering", func(t *trans.Transaction) {
		t.Request.URI = "/api/offerings/" + createdUUID
		t.FullPath = "/api/offerings/" + createdUUID
	})

	server.Serve()
	defer server.Listener.Close()
}
