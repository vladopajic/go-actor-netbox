package cp

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/vladopajic/go-actor/actor"
)

func NewConsumer(
	inMbx actor.MailboxReceiver[[]byte],
) actor.Actor {
	w := &consumerWorker{inMbx}
	return actor.New(w)
}

type consumerWorker struct {
	inMbx actor.MailboxReceiver[[]byte]
}

func (w *consumerWorker) DoWork(c actor.Context) actor.WorkerStatus {
	select {
	case <-c.Done():
		return actor.WorkerEnd

	case data := <-w.inMbx.ReceiveC():
		var num uint64
		err := binary.Read(bytes.NewReader(data), binary.BigEndian, &num)
		if err != nil {
			fmt.Printf("binary.Read failed: %v\n", err)
			return actor.WorkerContinue
		}

		fmt.Printf("consumed %v\n", num)

		return actor.WorkerContinue
	}
}
