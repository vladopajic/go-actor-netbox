package netbox

import (
	"github.com/gorilla/websocket"
	"github.com/vladopajic/go-actor/actor"
)

type WebsocketReceiver interface {
	actor.Actor
	actor.MailboxReceiver[[]byte]
	SetConn(conn *websocket.Conn)
}

type WebsocketSender interface {
	actor.Actor
	actor.MailboxSender[[]byte]
	SetConn(conn *websocket.Conn)
}

type msgPromise struct {
	msg  []byte
	errC chan error
}
