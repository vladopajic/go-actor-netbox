package netbox

import (
	"github.com/gorilla/websocket"
	"github.com/vladopajic/go-actor/actor"
)

type Receiver interface {
	actor.Actor
	actor.MailboxReceiver[[]byte]
}

type Sender interface {
	actor.Actor
	actor.MailboxSender[[]byte]
	SetConn(conn *websocket.Conn)
}
