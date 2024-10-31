package netbox

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/vladopajic/go-actor/actor"
)

func NewWsReceiver(conn *websocket.Conn) Receiver {
	receiver := &wsReceiver{
		mbx:  actor.NewMailbox[[]byte](),
		conn: conn,
	}
	receiver.Actor = actor.Combine(actor.New(receiver), receiver.mbx).Build()
	receiver.MailboxReceiver = receiver.mbx

	return receiver
}

type wsReceiver struct {
	actor.Actor
	actor.MailboxReceiver[[]byte]

	mbx  actor.Mailbox[[]byte]
	conn *websocket.Conn
}

func (r *wsReceiver) DoWork(ctx context.Context) actor.WorkerStatus {
	select {
	case <-ctx.Done():
		return actor.WorkerEnd
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

func NewWsSender() Sender {
	sender := &wsSender{
		mbx:   make(chan msgPromise),
		connC: make(chan *websocket.Conn),
	}
	sender.Actor = actor.New(sender)

	return sender
}

type wsSender struct {
	actor.Actor

	mbx   chan msgPromise
	connC chan *websocket.Conn
	conn  *websocket.Conn
}

func (s *wsSender) SetConn(conn *websocket.Conn) {
	s.connC <- conn
}

type msgPromise struct {
	msg  []byte
	errC chan error
}

func (s *wsSender) Send(ctx context.Context, msg []byte) error {
	msgProm := msgPromise{msg: msg, errC: make(chan error, 1)}

	select {
	case <-ctx.Done():
		return ctx.Err() //nolint:wrapcheck // relax
	case s.mbx <- msgProm:
	}

	return <-msgProm.errC
}

func (s *wsSender) DoWork(ctx context.Context) actor.WorkerStatus {
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

func (s *wsSender) handleSend(_ context.Context, msg []byte) error {
	return s.conn.WriteMessage(websocket.BinaryMessage, msg) //nolint:wrapcheck // relax
}
