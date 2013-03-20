package eventsocket

import (
	"bufio"
	"bytes"
	"log"
	"net/textproto"
	"strconv"
	"strings"
)

var (
	singleLine = []byte("\n")
	doubleLine = []byte("\n\n")
)

// Searches for next packet boundary
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

// Only PLAIN events supported at the moment
// They can be easily converted to json using encoding/json
func parseMessage(r *bufio.Reader) Event {
	var err error

	retMsg := new(eventReply)
	retMsg.headers, err = textproto.NewReader(r).ReadMIMEHeader()
	if err != nil {
		retMsg.realType = EventError
		return retMsg
	}

	bodyLenStr := retMsg.Headers().Get("Content-Length")

	if bodyLenStr != "" {
		//has body, go parse it

		bodyLen, err := strconv.Atoi(bodyLenStr)
		//content length with invalid size
		if err != nil {
			log.Printf("Failed to convert body")
			retMsg.success = false
			return retMsg
		}

		retMsg.body = make([]byte, bodyLen)

		read, err := r.Read(retMsg.body)

		if err != nil || read != bodyLen {
			retMsg.success = false
			return retMsg
		}
	}

	ctype := retMsg.Headers().Get("Content-Type")
	switch ctype {
	case "auth/request":
		retMsg.realType = EventAuth
	case "command/reply":
		retMsg.realType = EventReply
		replyText := retMsg.Headers().Get("Reply-Text")
		if strings.Contains(replyText, "+OK") {
			retMsg.success = true
		}
	case "text/event-plain", "text/event-json", "text/event-xml":
		retMsg.realType = EventGeneric
	case "text/disconnect-notice":
		retMsg.realType = EventDisconnect
	case "api/response":
		retMsg.realType = EventApi
		replyText := string(retMsg.Body())
		if strings.Contains(replyText, "+OK") {
			retMsg.success = true
		}
	}

	return retMsg
}
