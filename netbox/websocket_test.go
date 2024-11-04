package netbox_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/vladopajic/go-actor/actor"

	. "github.com/vladopajic/go-actor-netbox/netbox"
)

func Test_Websocket_WorkerEndSig(t *testing.T) {
	t.Parallel()

	r, s := NewWebsocketReceiver(), NewWebsocketSender()
	actor.AssertWorkerEndSig(t, r)
	actor.AssertWorkerEndSig(t, s)

	r.SetConn(&websocket.Conn{})
	actor.AssertWorkerEndSigAfterIterations(t, r, 2)

	s.SetConn(&websocket.Conn{})
	actor.AssertWorkerEndSigAfterIterations(t, s, 2)
}

func Test_Websocket_Integrated(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		return
	}

	ctx := context.Background()
	senderMbx := NewWebsocketSender()
	receiverMbx := NewWebsocketReceiver()

	doneC := make(chan any)

	a := actor.Combine(senderMbx, receiverMbx).Build()
	a.Start()
	defer a.Stop()

	go func() {
		http.HandleFunc("/ws", wsHandler(doneC, senderMbx.SetConn))

		//nolint:gosec // we don't care about timeouts here
		fatalErr(http.ListenAndServe(":8089", nil))
	}()

	time.Sleep(time.Second) //nolint:forbidigo // wait some time for server to start
	receiverMbx.SetConn(makeWsConn())

	go func() {
		for i := range 10 {
			fatalErr(senderMbx.Send(ctx, iToBytes(i)))
		}

		//nolint:contextcheck // relax
		assert.Error(t, senderMbx.Send(actor.ContextEnded(), iToBytes(100)))
	}()

	for i := range 10 {
		data := <-receiverMbx.ReceiveC()
		assert.Equal(t, i, iFromBytes(data))
	}

	close(doneC)
}

func wsHandler(doneC chan any, cb func(conn *websocket.Conn)) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool { return true },
	}

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		fatalErr(err)
		defer conn.Close()

		cb(conn)

		<-doneC
	}
}

func makeWsConn() *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: "localhost:8089", Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil) //nolint:bodyclose // relax
	fatalErr(err)

	return conn
}

func iToBytes(i int) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, uint64(i))
	fatalErr(err)

	return buf.Bytes()
}

func iFromBytes(bb []byte) int {
	var num uint64
	err := binary.Read(bytes.NewReader(bb), binary.BigEndian, &num)
	fatalErr(err)

	return int(num) //nolint:gosec // we don't care about overflow here
}

func fatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
