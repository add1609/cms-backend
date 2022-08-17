package main

import (
	"cms-backend/src"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

func main() {
	myEnv, err := godotenv.Read(".env.local")
	if err != nil {
		log.Fatalf("[FATAL ERROR] Loading .env.local file in main(): %v", err)
	}

	hub := src.NewHub(myEnv)
	go hub.Run()
	http.HandleFunc("/socket", func(w http.ResponseWriter, r *http.Request) {
		src.ServeWs(hub, w, r)
	})

	if myEnv["SETUP_TLS"] == "true" {
		src.SetupTLS(
			myEnv["CONFIG_DOMAIN"],
			myEnv["CONFIG_CACHE_DIR"],
			myEnv["CONFIG_SSL_EMAIL"],
			myEnv["CONFIG_HTTP_ADDRESS"],
			myEnv["WS_CHECK_ORIGIN_HOST"],
		)
	} else {
		log.Printf("[INFO] Starting HTTP listener on port %s", myEnv["WS_PORT"])
		err = http.ListenAndServe(":"+myEnv["WS_PORT"], nil)
		if err != nil {
			log.Fatalf("[FATAL ERROR] in ListenAndServe() on port %s: %v", myEnv["WS_PORT"], err)
		}
	}
}
