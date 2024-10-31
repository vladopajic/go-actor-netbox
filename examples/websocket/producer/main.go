package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/vladopajic/go-actor-netbox/examples/cp"
	"github.com/vladopajic/go-actor-netbox/netbox"
	"github.com/vladopajic/go-actor/actor"
)

func main() {
	senderMbx := netbox.NewWsSender()
	producer := actor.New(cp.NewProducerWorker(senderMbx))

	a := actor.Combine(senderMbx, producer).Build()
	a.Start()
	defer a.Stop()

	http.HandleFunc("/ws", wsHandler(senderMbx))
	log.Fatal(http.ListenAndServe(":8088", nil))
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsHandler(senderMbx netbox.Sender) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		defer conn.Close()

		senderMbx.SetConn(conn)

		select {}
	}
}
