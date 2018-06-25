package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/WasinWatt/slumbot/config"
	"github.com/WasinWatt/slumbot/postgres"
	"github.com/WasinWatt/slumbot/service"
	"github.com/jinzhu/configor"

	"github.com/WasinWatt/slumbot/api"
	_ "github.com/lib/pq"
	"github.com/line/line-bot-sdk-go/linebot"
)

func main() {
	conf := &config.Config{}
	configor.Load(conf, "config.yml")

	bot, err := linebot.New(conf.ChannelSecret, conf.ChannelAccToken)
	must(err)

	db, err := sql.Open("postgres", conf.PostgresURI)
	must(err)

	log.Println("Connected to DB ...")
	// Repo initialize
	repo := postgres.New()

	// Service controller initialize
	controller := service.New(db, repo)

	apiHandler := api.NewHandler(bot, db, controller)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.NotFound(w, req)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response, _ := json.Marshal("Status OK")
		w.WriteHeader(200)
		w.Write(response)
	})

	mux.Handle("/api/", http.StripPrefix("/api", apiHandler.MakeHandler()))

	must(err)

	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "3000"
	}

	log.Println("Listening on port: " + addr)
	http.ListenAndServe(":"+addr, mux)

}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
