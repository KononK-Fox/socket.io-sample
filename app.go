package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	socketio "github.com/googollee/go-socket.io"
)

var (
	users map[string]string = make(map[string]string)
)

type PayloadJoin struct {
	Name string `json:"name"`
}

type PayloadChat struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	To      string `json:"to"`
}

func main() {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", func(so socketio.Socket) {
		so.Join("main_room")
		so.Join("private_room" + so.Id())

		so.On("event_join", func(message string) {
			payload := PayloadJoin{}
			json.Unmarshal([]byte(message), &payload)

			users[strings.ToLower(payload.Name)] = so.Id()

			so.BroadcastTo("main_room", "event_join", message)
		})

		so.On("event_chat", func(message string) {
			payload := PayloadChat{}
			json.Unmarshal([]byte(message), &payload)

			if payload.To != "" {
				so.BroadcastTo("private_room"+users[strings.ToLower(payload.To)], "event_chat", message)
			} else {
				so.BroadcastTo("main_room", "event_chat", message)
			}

		})
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content, err := os.ReadFile("index.html")
		if err != nil {
			http.Error(w, "Could not open requested file", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%s", content)
	})

	http.Handle("/socket.io/", server)

	fmt.Println("Server starting at :8080")
	http.ListenAndServe(":8080", nil)
}
