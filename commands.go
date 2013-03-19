package eventsocket

import (
	"bytes"
)

func (conn *FSConnection) EventPlain(event string, moreevents ...string) {
	cmd := bytes.NewBufferString("event plain ")
	cmd.WriteString(event)
	for _, v := range moreevents {
		cmd.WriteString(" ")
		cmd.WriteString(v)
	}
	cmd.Write(doubleLine)
	conn.send(cmd.Bytes())
}
