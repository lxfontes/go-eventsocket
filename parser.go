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
func parseMessage(r *bufio.Reader) *FSMessage {
	var err error
	var retMsg = new(FSMessage)
	retMsg.Headers, err = textproto.NewReader(r).ReadMIMEHeader()
	if err != nil {
		retMsg.Type = ParserError
		return retMsg
	}

	ctype := retMsg.Headers.Get("Content-Type")
	switch ctype {
	case "auth/request":
		retMsg.Type = RequestAuthentication
	case "command/reply":
		retMsg.Type = CommandReply
	case "text/event-plain":
		retMsg.Type = EventPlain
	case "text/disconnect-notice":
		retMsg.Type = DisconnectNotice
	}

	replyText := retMsg.Headers.Get("Reply-Text")

	if strings.Contains(replyText, "+OK") {
		retMsg.Success = true
	}

	bodyLenStr := retMsg.Headers.Get("Content-Length")

	if bodyLenStr != "" {
		//has body, go parse it
		bodyLen, err := strconv.Atoi(bodyLenStr)
		//content length with invalid size
		if err != nil {
			log.Printf("Failed to convert body")
			retMsg.Success = false
			return retMsg
		}
		log.Printf("Bodylen %d", bodyLen)

		bodyb, err := seeUpcomingHeader(r)
		if !bodyb {
			retMsg.Success = false
			return retMsg
		} else {
			retMsg.Body, err = textproto.NewReader(r).ReadMIMEHeader()
			if err != nil {
				retMsg.Success = false
				return retMsg
			}
		}
	}

	return retMsg
}
