package eventsocket

import (
	"bufio"
	"bytes"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
)

var (
	singleLine = []byte("\n")
	doubleLine = []byte("\n\n")
)

const BufferSize = 16 * 1024

type Command struct {
	Lock bool
	Uuid string
	App  string
	Args string
}

type ESLkv struct {
	body         textproto.MIMEHeader
	shouldEscape bool
}

// EventType indicates how/why this event is being raised
type EventType int

const (
	//Parser and Socket errors
	EventError EventType = iota
	//Client Mode: Connection established / terminated
	//Server Mode: New Connection / Lost Connection
	EventState
	//Library internal, used during auth phase
	EventAuth
	//Command Reply
	EventReply
	//API (reloadxml,reloadacl) commands
	EventApi
	//Received a disconnect notice (linger might still be on)
	EventDisconnect
	//Subscribed events
	EventGeneric
)

type Event struct {
	Type      EventType
	Headers   ESLkv
	Body      []byte
	EventBody ESLkv
	Success   bool
}

// Searches for next packet boundary
// Blocks until we can 'peek' inside the reader
func seeUpcomingHeader(r *bufio.Reader) (bool, error) {
	for peekSize := 4; ; peekSize++ {
		// This loop stops when Peek returns an error,
		// which it does when r's buffer has been filled.
		buf, err := r.Peek(peekSize)
		if bytes.HasSuffix(buf, doubleLine) {
			return true, nil
		}
		if err != nil {
			return false, err
			break
		}
	}
	return false, nil
}

// Parses header and body
// Body can be easily converted to json using encoding/json
func parseMessage(r *bufio.Reader) *Event {
	var err error

	retMsg := new(Event)
	retMsg.Headers.body, err = textproto.NewReader(r).ReadMIMEHeader()
	if err != nil {
		retMsg.Type = EventError
		return retMsg
	}

	bodyLenStr := retMsg.Headers.Get("Content-Length")

	if len(bodyLenStr) > 0 {
		//has body, go parse it

		bodyLen, err := strconv.Atoi(bodyLenStr)
		//content length with invalid size
		if err != nil {
			retMsg.Success = false
			return retMsg
		}

		retMsg.Body = make([]byte, bodyLen)

		_, err = r.Peek(bodyLen)
		if err != nil {
			retMsg.Success = false
			return retMsg
		}

		read, err := r.Read(retMsg.Body)

		if err != nil || read != bodyLen {
			retMsg.Success = false
			return retMsg
		}
	}

	ctype := retMsg.Headers.Get("Content-Type")

	switch ctype {
	case "auth/request":
		retMsg.Type = EventAuth
	case "command/reply":
		retMsg.Type = EventReply
		replyText := retMsg.Headers.Get("Reply-Text")
		if strings.Contains(replyText, "+OK") {
			retMsg.Success = true
		}
		if strings.Contains(replyText, "%") {
			retMsg.Headers.shouldEscape = true
		}
	case "text/event-plain":
		retMsg.Type = EventGeneric
		retMsg.EventBody.shouldEscape = true
		retMsg.parseBody()
	case "text/event-json", "text/event-xml":
		retMsg.Type = EventGeneric
	case "text/disconnect-notice":
		retMsg.Type = EventDisconnect
	case "api/response":
		retMsg.Type = EventApi
		replyText := string(retMsg.Body)
		if strings.Contains(replyText, "+OK") {
			retMsg.Success = true
		}
	}

	return retMsg
}

func (evt *Event) parseBody() {
	var err error
	evt.EventBody.body, err = textproto.NewReader(bufio.NewReader(bytes.NewBuffer(evt.Body))).ReadMIMEHeader()
	if err != nil {
		evt.Success = false
		return
	}
}

func localUnescape(s string) string {
	x, _ := url.QueryUnescape(s)
	return x
}

func (eB *ESLkv) Get(key string) string {
	s := eB.body.Get(key)
	if eB.shouldEscape {
		return localUnescape(s)
	}
	return s
}

// readMessage blocks until a message is available
func readMessage(rw *bufio.ReadWriter) *Event {
	for {
		headerReady, err := seeUpcomingHeader(rw.Reader)

		if err != nil {
			break
		}

		if headerReady {
			//good to parse
			message := parseMessage(rw.Reader)
			return message
		}
	}

	//returns a zeroed event, which is an error
	return new(Event)
}

func sendBytes(rw *bufio.ReadWriter, b []byte) (int, error) {
	defer rw.Flush()
	return rw.Write(b)
}

// GetExecute formats a dialplan command to be sent over TCP Connection
func (cmd *Command) GetExecute() []byte {
	bbuf := bytes.NewBufferString("sendmsg")
	if len(cmd.Uuid) > 0 {
		bbuf.WriteString(" ")
		bbuf.WriteString(cmd.Uuid)
	}
	bbuf.Write(singleLine)
	bbuf.WriteString("call-command: execute")
	bbuf.Write(singleLine)
	bbuf.WriteString("execute-app-name: ")
	bbuf.WriteString(cmd.App)
	bbuf.Write(singleLine)
	if len(cmd.Args) > 0 {
		bbuf.WriteString("execute-app-arg: ")
		bbuf.WriteString(cmd.Args)
		bbuf.Write(singleLine)
	}
	if cmd.Lock {
		bbuf.WriteString("event-lock: true")
		bbuf.Write(singleLine)
	}
	bbuf.Write(doubleLine)
	return bbuf.Bytes()
}
