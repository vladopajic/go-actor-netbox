package netbox

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/vladopajic/go-actor/actor"
)

func NewWebsocketReceiver() WebsocketReceiver {
	receiver := &websocketReceiver{
		mbx:   actor.NewMailbox[[]byte](),
		connC: make(chan *websocket.Conn, 1),
	}
	receiver.Actor = actor.Combine(actor.New(receiver), receiver.mbx).Build()
	receiver.MailboxReceiver = receiver.mbx

	return receiver
}

type websocketReceiver struct {
	actor.Actor
	actor.MailboxReceiver[[]byte]

	mbx   actor.Mailbox[[]byte]
	connC chan *websocket.Conn
	conn  *websocket.Conn
}

func (r *websocketReceiver) SetConn(conn *websocket.Conn) {
	r.connC <- conn
}

func (r *websocketReceiver) DoWork(ctx context.Context) actor.WorkerStatus {
	if r.conn == nil {
		select {
		case <-ctx.Done():
			return actor.WorkerEnd

		case conn := <-r.connC:
			r.conn = conn
			return actor.WorkerContinue
		}
	}

	select {
	case <-ctx.Done():
		return actor.WorkerEnd

	case conn := <-r.connC:
		r.conn = conn
		return actor.WorkerContinue

	default:
	}

	msgType, message, err := r.conn.ReadMessage()
	if err != nil {
		return actor.WorkerEnd
	}

	if msgType == websocket.BinaryMessage {
		r.mbx.Send(ctx, message) //nolint:errcheck // should never err
	}

	return actor.WorkerContinue
}

func NewWebsocketSender() WebsocketSender {
	sender := &websocketSender{
		mbx:   make(chan msgPromise),
		connC: make(chan *websocket.Conn, 1),
	}
	sender.Actor = actor.New(sender)

	return sender
}

type websocketSender struct {
	actor.Actor

	mbx   chan msgPromise
	connC chan *websocket.Conn
	conn  *websocket.Conn
}

func (s *websocketSender) SetConn(conn *websocket.Conn) {
	s.connC <- conn
}

func (s *websocketSender) Send(ctx context.Context, msg []byte) error {
	msgProm := msgPromise{msg: msg, errC: make(chan error, 1)}

	select {
	case <-ctx.Done():
		return ctx.Err() //nolint:wrapcheck // relax
	case s.mbx <- msgProm:
	}

	select {
	case <-ctx.Done():
		return ctx.Err() //nolint:wrapcheck // relax
	case err := <-msgProm.errC:
		return err
	}
}

func (s *websocketSender) DoWork(ctx context.Context) actor.WorkerStatus {
	if s.conn == nil {
		select {
		case <-ctx.Done():
			return actor.WorkerEnd

		case conn := <-s.connC:
			s.conn = conn
			return actor.WorkerContinue
		}
	}

	select {
	case <-ctx.Done():
		return actor.WorkerEnd

	case conn := <-s.connC:
		s.conn = conn
		return actor.WorkerContinue

	case msgProm := <-s.mbx:
		msgProm.errC <- s.handleSend(ctx, msgProm.msg)

		return actor.WorkerContinue
	}
}

func (s *websocketSender) handleSend(_ context.Context, msg []byte) error {
	return s.conn.WriteMessage(websocket.BinaryMessage, msg) //nolint:wrapcheck // relax
}
