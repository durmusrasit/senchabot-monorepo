package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gempir/go-twitch-irc/v3"
	"github.com/joho/godotenv"
	"github.com/senchabot-dev/monorepo/apps/twitch-bot/client"
	"github.com/senchabot-dev/monorepo/apps/twitch-bot/internal/backend/mysql"
	"github.com/senchabot-dev/monorepo/apps/twitch-bot/internal/db"
	"github.com/senchabot-dev/monorepo/apps/twitch-bot/internal/handler"
	"github.com/senchabot-dev/monorepo/apps/twitch-bot/server"
)

func HandleBotJoinWebhook(twitch *twitch.Client, joinedChannelList []string, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var data struct {
		Event string `json:"event"`
		User  string `json:"user_name"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}
	fmt.Println("data", data)
	channel := strings.TrimPrefix(data.Event, "channel.join.")

	// check if channel is not in joinedChannelList
	for _, v := range joinedChannelList {
		if v == channel {
			return
			// http.Error(w, "Already joined", http.StatusBadRequest)
			// return
		}
	}

	joinedChannelList = append(joinedChannelList, channel)
	twitch.Join(channel)

	w.WriteHeader(http.StatusOK)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	twitchClient := twitch.NewClient("senchabot", os.Getenv("OAUTH"))

	mySQLBackend := mysql.NewMySQLBackend(db.NewMySQL())
	server := server.NewSenchabotAPIServer(mySQLBackend)

	clients := client.NewClients(twitchClient)

	joinedChannelList := handler.InitHandlers(clients, server)

	go func() {
		fmt.Println("Connecting to Twitch...")
		error := twitchClient.Connect()
		if error != nil {
			panic("Connecting to Twitch Error" + error.Error())
		}
	}()

	go func() {
		fmt.Println("Starting HTTP server...")
		mux := http.NewServeMux()
		mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
			HandleBotJoinWebhook(twitchClient, joinedChannelList, w, r)
		})
		error := http.ListenAndServe(":8080", mux)
		if error != nil {
			log.Fatal("ListenAndServe Error:", error)
		}
	}()

	select {}
}
