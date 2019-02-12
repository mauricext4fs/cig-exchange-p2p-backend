package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"github.com/snikch/goodman/hooks"
	trans "github.com/snikch/goodman/transaction"
)

func main() {

	h := hooks.NewHooks()
	server := hooks.NewServer(hooks.NewHooksRunner(h))

	err := godotenv.Load()
	if err != nil {
		fmt.Print(err)
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	client := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err = client.Ping().Result()
	if err != nil {
		log.Print(err.Error())
	}

	// save created record UUID here
	createdUUID := ""

	// user JWT saved by homepage dredd tests
	userJWT := ""
	jwtCmd := client.Get("jwt")
	if jwtCmd.Err() != nil {
		log.Printf("Unable to get jwt from redis: Error: %v", jwtCmd.Err().Error())
		return
	}
	userJWT = jwtCmd.Val()

	// inject JWT auth into all calls
	h.BeforeEach(func(t *trans.Transaction) {
		if len(userJWT) == 0 || t.Request == nil {
			return
		}

		t.Request.Headers["Authorization"] = fmt.Sprintf("Bearer %v", userJWT)
	})

	h.After("Offerings > p2p/api/offerings > Create offering", func(t *trans.Transaction) {
		// happens when api is down
		if t.Real == nil {
			return
		}

		body := map[string]interface{}{}

		json.Unmarshal([]byte(t.Real.Body), &body)
		createdUUIDInterface, ok := body["id"]
		if !ok {
			log.Printf("Error: Unable to save the uuid of the created offering")
			return
		}
		createdUUID = createdUUIDInterface.(string)
	})

	// update URI everywhere to point to a created record
	h.Before("Offerings > p2p/api/offerings/{offering_id} > Retrieve offering", func(t *trans.Transaction) {
		if len(createdUUID) > 0 {
			t.Request.URI = "/p2p/api/offerings/" + createdUUID
			t.FullPath = "/p2p/api/offerings/" + createdUUID
		}
	})

	h.Before("Offerings > p2p/api/offerings/{offering_id} > Update offering", func(t *trans.Transaction) {
		if len(createdUUID) > 0 {
			t.Request.URI = "/p2p/api/offerings/" + createdUUID
			t.FullPath = "/p2p/api/offerings/" + createdUUID
		}
	})

	h.Before("Offerings > p2p/api/offerings/{offering_id} > Delete offering", func(t *trans.Transaction) {
		if len(createdUUID) > 0 {
			t.Request.URI = "/p2p/api/offerings/" + createdUUID
			t.FullPath = "/p2p/api/offerings/" + createdUUID
		}
	})

	server.Serve()
	defer server.Listener.Close()
}
