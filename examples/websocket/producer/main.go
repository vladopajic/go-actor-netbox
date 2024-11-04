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
	senderMbx := netbox.NewWebsocketSender()
	producer := cp.NewProducer(senderMbx)

	a := actor.Combine(senderMbx, producer).Build()
	a.Start()
	defer a.Stop()

	http.HandleFunc("/ws", wsHandler(senderMbx.SetConn))
	log.Fatal(http.ListenAndServe(":8088", nil))
}

func wsHandler(cb func(conn *websocket.Conn)) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool { return true },
	}

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		defer conn.Close()

		cb(conn)

		select {}
	}
}
