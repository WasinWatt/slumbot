package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/WasinWatt/slumbot/cache"
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

	// cache initialize
	memcache := cache.New()
	// Repo initialize
	repo := postgres.New(memcache)

	// Service controller initialize
	controller := service.New(db, repo)

	apiHandler := api.NewHandler(bot, db, controller, memcache)

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
		addr = "5000"
	}

	server := &http.Server{
		Addr:    ":" + addr,
		Handler: mux,
	}

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Kill, os.Interrupt)

		<-sig
		log.Println("Catch signal ... Shutting down the server")
		err := server.Shutdown(context.Background())
		if err != nil {
			log.Println("Fail while shutting down the server")
		} else {
			log.Println("Server stopped")
		}
	}()

	log.Println("Listening on port: " + addr)
	server.ListenAndServe()

}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
