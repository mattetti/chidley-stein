package chidleystein

import (
	"fmt"
)

type StdoutWriter struct {
}

var doneChannel chan bool

func (w *StdoutWriter) Open(s string, lineChannel chan string) error {
	doneChannel = make(chan bool)
	go w.Writer(lineChannel, doneChannel)

	return nil
}

func (w *StdoutWriter) Writer(lineChannel chan string, doneChannel chan bool) {
	for line := range lineChannel {
		fmt.Println(line)
	}
	doneChannel <- true
}

func (w *StdoutWriter) Close() {
	_ = <-doneChannel
}
