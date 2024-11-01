package cp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/vladopajic/go-actor/actor"
)

func NewProducer(
	outMbx actor.MailboxSender[[]byte],
) actor.Actor {
	w := &producerWorker{outMbx: outMbx}
	return actor.New(w)
}

type producerWorker struct {
	outMbx actor.MailboxSender[[]byte]
	num    uint64
}

func (w *producerWorker) DoWork(c actor.Context) actor.WorkerStatus {
	select {
	case <-c.Done():
		return actor.WorkerEnd

	case <-time.After(time.Second):
		w.num++

		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.BigEndian, w.num)
		if err != nil {
			fmt.Printf("binary.Write failed: %v\n", err)
			return actor.WorkerContinue
		}

		err = w.outMbx.Send(c, buf.Bytes())
		if err != nil {
			fmt.Printf("outMbx.Send failed: %v\n", err)
			return actor.WorkerContinue
		}

		fmt.Print(".")

		return actor.WorkerContinue
	}
}
