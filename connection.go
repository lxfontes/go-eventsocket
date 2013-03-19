/*
Freeswitch Event Socket bindings
*/
package eventsocket

import (
	"bufio"
	"bytes"
	"net"
	"time"
)

func CreateClient(address string, password string, handler FSHandler, timeout time.Duration) (*FSListener, error) {
	var retClient = new(FSListener)
	retClient.Address = address
	retClient.ClientMode = true
	retClient.Handler = handler
	retClient.Password = password
	retClient.Timeout = timeout
	return retClient, nil
}

func (conn *FSConnection) tryConnect() {
	//check if we should connect
	canConnect := conn.Listener.Handler.CreateConnection(conn)
	var err error

	if canConnect {
		conn.Connection, err = net.DialTimeout("tcp", conn.Listener.Address, conn.Listener.Timeout)

		if err != nil {
			time.Sleep(conn.Listener.Timeout)
			return
		}

	} else {
		// handler denied connection, just wait
		time.Sleep(conn.Listener.Timeout)
		return
	}

	//Got connection, initiate authentication
	conn.rw = bufio.NewReadWriter(
		bufio.NewReader(conn.Connection),
		bufio.NewWriter(conn.Connection))

	//wait for authentication request
	auth := conn.ReadMessage()

	if auth.Type != RequestAuthentication {
		//invalid handshake
		return
	}

	//send password
	pwBuf := bytes.NewBufferString("auth ")
	pwBuf.WriteString(conn.Listener.Password)
	pwBuf.WriteString("\n\n")
	conn.rw.Write(pwBuf.Bytes())
	err = conn.rw.Flush()

	if err != nil {
		conn.Connection.Close()
		return
	}

	authResp := conn.ReadMessage()

	if authResp.Type != CommandReply {
		conn.Connection.Close()
		return
	}

	if !authResp.Success {
		conn.Connection.Close()
		return
	}

	conn.Connected = true
	conn.Listener.Handler.ConnectionAccepted(conn)
}

// Loop communicates with main process via Handler's message channel
// To be called after authentication or new connection
func (listener *FSListener) loopClient() {
	var fscon = new(FSConnection)
	fscon.Listener = listener

	for {

		// Handles reconnect as needed
		if !fscon.Connected {
			fscon.tryConnect()
			continue
		}

		message := fscon.ReadMessage()
		if message.Type == ParserError {
			//disconnect
			fscon.Connection.Close()
			fscon.Connected = false
			listener.Handler.CloseConnection(fscon)
		} else {
			listener.Handler.HandleEvent(fscon, message)
		}

	}

}

func (listener *FSListener) loopServer() {

}

func (listener *FSListener) Loop() {
	if listener.ClientMode {
		listener.loopClient()
	} else {
		listener.loopServer()
	}
}

// ReadMessage blocks until a message is available
func (connection *FSConnection) ReadMessage() *FSMessage {
	for {
		headerReady, err := seeUpcomingHeader(connection.rw.Reader)

		if err != nil {
			break
		}

		if headerReady {
			//good to parse
			message := parseMessage(connection.rw.Reader)
			return message
		}
	}
	return new(FSMessage)
}

func (connection *FSConnection) send(b []byte) (int, error) {
	defer connection.rw.Flush()
	return connection.rw.Write(b)
}

// Message Handling

func (msg *FSMessage) String() string {
	return "mymessage"
}
