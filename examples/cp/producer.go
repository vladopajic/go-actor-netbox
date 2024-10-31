package cp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/vladopajic/go-actor/actor"
)

func NewProducerWorker(
	outMbx actor.MailboxSender[[]byte],
) actor.Worker {
	return &producerWorker{
		outMbx: outMbx,
	}
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

		w.outMbx.Send(c, buf.Bytes())

		return actor.WorkerContinue
	}
}
