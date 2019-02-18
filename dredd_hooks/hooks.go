package main

import (
	"cig-exchange-libs/models"
	"encoding/json"
	"fmt"

	"cig-exchange-libs"

	"github.com/snikch/goodman/hooks"
	trans "github.com/snikch/goodman/transaction"
)

const dredd = "dredd"
const dreddP2P = "dredd_p2p"

func main() {

	h := hooks.NewHooks()
	server := hooks.NewServer(hooks.NewHooksRunner(h))

	redisClient := cigExchange.GetRedis()

	// save created record UUID here
	createdUUID := ""
	// user JWT saved by homepage dredd tests
	userJWT := ""
	// user organisation UUID saved by homepage dredd tests
	orgUUID := ""

	jwtCmd := redisClient.Get("jwt")
	if jwtCmd.Err() != nil {
		fmt.Println("ERROR: Unable to get jwt from redis:")
		fmt.Println(jwtCmd.Err().Error())
	} else {
		userJWT = jwtCmd.Val()
	}

	orgCmd := redisClient.Get("org")
	if orgCmd.Err() != nil {
		fmt.Println("ERROR: Unable to get organisation UUID from redis:")
		fmt.Println(orgCmd.Err().Error())
	} else {
		orgUUID = orgCmd.Val()
	}

	// inject JWT auth into all calls
	h.BeforeEach(func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}

		if len(userJWT) == 0 {
			t.Fail = "JWT missing"
		}

		t.Request.Headers["Authorization"] = fmt.Sprintf("Bearer %v", userJWT)
	})

	// update URI everywhere to point to a created record
	h.Before("Offerings > p2p/api/organisations/{organisation_id}/offerings > Create offering", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/offerings"
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/offerings"
		setBodyValue(&t.Request.Body, "title", dreddP2P)
		setBodyValue(&t.Request.Body, "organisation_id", orgUUID)
	})

	h.After("Offerings > p2p/api/organisations/{organisation_id}/offerings > Create offering", func(t *trans.Transaction) {
		// happens when api is down
		if t.Real == nil {
			return
		}

		createdUUID = getBodyValue(&t.Real.Body, "id")
		if len(createdUUID) == 0 {
			t.Fail = "Unable to save the uuid of the created offering"
			return
		}
	})

	h.Before("Offerings > p2p/api/organisations/{organisation_id}/offerings > Retrieve all offerings", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/offerings"
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/offerings"
	})

	h.After("Offerings > p2p/api/organisations/{organisation_id}/offerings > Retrieve all offerings", func(t *trans.Transaction) {
		// happens when api is down
		if t.Real == nil {
			return
		}

		// verify that created offerings are present in 'p2p/offerings' api call
		offerings := make([]models.Offering, 0)
		err := json.Unmarshal([]byte(t.Real.Body), &offerings)
		if err != nil {
			t.Fail = fmt.Sprintf("Unable to parse response: %v", err.Error())
			return
		}

		p2pFound := false
		homepageFound := false
		for _, offering := range offerings {
			if offering.Title == dredd {
				homepageFound = true
				continue
			}

			if offering.Title == dreddP2P {
				p2pFound = true
				continue
			}
		}

		if p2pFound && homepageFound {
			// we found a match, api works fine
			return
		}

		t.Fail = "Pre-created offering is missing"
	})

	h.Before("Offerings > p2p/api/organisations/{organisation_id}/offerings/{offering_id} > Retrieve offering", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}
		if len(createdUUID) == 0 {
			t.Fail = "Created offering UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/offerings/" + createdUUID
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/offerings/" + createdUUID
	})

	h.Before("Offerings > p2p/api/organisations/{organisation_id}/offerings/{offering_id} > Update offering", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}
		if len(createdUUID) == 0 {
			t.Fail = "Created offering UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/offerings/" + createdUUID
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/offerings/" + createdUUID

		setBodyValue(&t.Request.Body, "organisation_id", orgUUID)
	})

	h.Before("Offerings > p2p/api/organisations/{organisation_id}/offerings/{offering_id} > Delete offering", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}
		if len(createdUUID) == 0 {
			t.Fail = "Created offering UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/offerings/" + createdUUID
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/offerings/" + createdUUID
	})

	server.Serve()
	defer server.Listener.Close()
}

func setBodyValue(body *string, key, value string) {

	if body == nil {
		return
	}

	bodyMap := map[string]interface{}{}

	err := json.Unmarshal([]byte(*body), &bodyMap)
	if err != nil {
		return
	}

	bodyMap[key] = value
	b, err := json.Marshal(bodyMap)
	if err != nil {
		return
	}

	*body = string(b)
}

func getBodyValue(body *string, key string) (value string) {

	if body == nil {
		return
	}

	bodyMap := map[string]interface{}{}

	err := json.Unmarshal([]byte(*body), &bodyMap)
	if err != nil {
		return
	}

	v, ok := bodyMap[key]
	if ok {
		// make sure it's a string
		vs, ok := v.(string)
		if ok {
			value = vs
		}
	}

	return
}
