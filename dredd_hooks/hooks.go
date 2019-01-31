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

	// save created user UUID and JWT
	userUUID := ""
	userJWT := ""

	h.After("Users > p2p/api/users/signin > Signin", func(t *trans.Transaction) {
		// happens when api is down
		if t.Real == nil {
			return
		}

		body := map[string]interface{}{}

		err := json.Unmarshal([]byte(t.Real.Body), &body)
		if err != nil {
			log.Printf("Signin: Error: %v", err.Error())
			return
		}

		createdUUIDInterface, ok := body["uuid"]
		if !ok {
			log.Printf("Signin: Error: Unable to save signin uuid")
			return
		}
		userUUID = createdUUIDInterface.(string)
	})

	h.Before("Users > p2p/api/users/send_otp > Send OTP", func(t *trans.Transaction) {

		if len(userUUID) == 0 || t.Request == nil {
			return
		}

		body := map[string]interface{}{}

		err := json.Unmarshal([]byte(t.Request.Body), &body)
		if err != nil {
			log.Printf("Send OTP: Error: %v", err.Error())
			return
		}

		body["uuid"] = userUUID
		b, err := json.Marshal(body)
		if err != nil {
			log.Printf("Send OTP: Error: %v", err.Error())
			return
		}

		t.Request.Body = string(b)
	})

	h.Before("Users > p2p/api/users/verify_otp > Verify OTP", func(t *trans.Transaction) {

		if len(userUUID) == 0 || t.Request == nil {
			return
		}

		body := map[string]interface{}{}

		err := json.Unmarshal([]byte(t.Request.Body), &body)
		if err != nil {
			log.Printf("Verify OTP: Error: %v", err.Error())
			return
		}

		rediskey := fmt.Sprintf("%s_signup_key", userUUID)
		redisCmd := client.Get(rediskey)
		if redisCmd.Err() != nil {
			log.Printf("Verify OTP: Error: %v", redisCmd.Err().Error())
			return
		}

		body["uuid"] = userUUID
		body["code"] = redisCmd.Val()
		b, err := json.Marshal(body)
		if err != nil {
			log.Printf("Verify OTP: Error: %v", err.Error())
			return
		}

		t.Request.Body = string(b)
	})

	h.After("Users > p2p/api/users/verify_otp > Verify OTP", func(t *trans.Transaction) {

		// happens when api is down
		if t.Real == nil {
			return
		}

		body := map[string]interface{}{}

		err := json.Unmarshal([]byte(t.Real.Body), &body)
		if err != nil {
			log.Printf("Verify OTP: Error: %v", err.Error())
			return
		}
		userJWTInterface, ok := body["jwt"]
		if !ok {
			log.Printf("Verify OTP: Error: Unable to save user jwt")
			return
		}
		userJWT = userJWTInterface.(string)
		log.Printf("jwt: %v", userJWT)
	})

	h.After("Offerings > api/offerings > Create offering", func(t *trans.Transaction) {
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
	h.Before("Offerings > api/offerings/{offering_id} > Retrieve offering", func(t *trans.Transaction) {
		if len(createdUUID) > 0 {
			t.Request.URI = "/api/offerings/" + createdUUID
			t.FullPath = "/api/offerings/" + createdUUID
		}
	})

	h.Before("Offerings > api/offerings/{offering_id} > Update offering", func(t *trans.Transaction) {
		if len(createdUUID) > 0 {
			t.Request.URI = "/api/offerings/" + createdUUID
			t.FullPath = "/api/offerings/" + createdUUID
		}
	})

	h.Before("Offerings > api/offerings/{offering_id} > Delete offering", func(t *trans.Transaction) {
		if len(createdUUID) > 0 {
			t.Request.URI = "/api/offerings/" + createdUUID
			t.FullPath = "/api/offerings/" + createdUUID
		}
	})

	server.Serve()
	defer server.Listener.Close()
}
