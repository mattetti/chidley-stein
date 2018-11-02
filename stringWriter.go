package chidleystein

var (
	_ Writer = (*StringWriter)(nil) // compilation check
)

type StringWriter struct {
	S string
}

func (w *StringWriter) Open(s string, lineChannel chan string) error {
	doneChannel = make(chan bool)
	go w.Writer(lineChannel, doneChannel)
	return nil
}

func (w *StringWriter) Writer(lineChannel chan string, doneChannel chan bool) {
	for line := range lineChannel {
		w.S += line + "\n"
	}
	doneChannel <- true
}

func (w *StringWriter) Close() {
	_ = <-doneChannel
}
