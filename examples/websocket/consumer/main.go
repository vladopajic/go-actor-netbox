package main

import (
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vladopajic/go-actor/actor"

	"github.com/vladopajic/go-actor-netbox/examples/cp"
	"github.com/vladopajic/go-actor-netbox/netbox"
)

func main() {
	conn := makeConn()
	defer conn.Close()

	receiverMbx := netbox.NewWebsocketReceiver()
	consumer := cp.NewConsumer(receiverMbx)

	a := actor.Combine(receiverMbx, consumer).Build()
	a.Start()
	defer a.Stop()

	receiverMbx.SetConn(conn)

	select {}
}

func makeConn() *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: "localhost:8088", Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial error:", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	return conn
}
