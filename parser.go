package eventsocket

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/textproto"
	"strconv"
	"strings"
)

var (
	singleLine = []byte("\n")
	doubleLine = []byte("\n\n")
)

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
	retMsg.Headers, err = textproto.NewReader(r).ReadMIMEHeader()
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
	case "text/event-plain", "text/event-json", "text/event-xml":
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

func ParseJson(evt *Event) (JsonBody, error) {
	var data JsonBody
	err := json.Unmarshal(evt.Body, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (con *connection) Send(cmd string, args ...string) (*Event, error) {
	bbuf := bytes.NewBufferString(cmd)
	for _, item := range args {
		bbuf.WriteString(" ")
		bbuf.WriteString(item)
	}
	bbuf.Write(doubleLine)

	//all commands block until FS replies
	con.lock.Lock()
	defer con.lock.Unlock()

	sendBytes(con.rw, bbuf.Bytes())
	evt := <-con.apiChan

	return evt, nil
}

func (con *connection) Execute(cmd *Command) (*Event, error) {
	con.lock.Lock()
	defer con.lock.Unlock()

	sendBytes(con.rw, cmd.GetExecute())
	evt := <-con.apiChan

	return evt, nil
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
