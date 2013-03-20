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

func CreateClient(settings *ClientSettings) (*Client, error) {
	retClient := new(Client)
	retClient.Settings = settings
	retClient.executeChan = make(chan Event)
	retClient.apiChan = make(chan Event)
	go retClient.Loop()
	return retClient, nil
}

func (client *Client) Close() {
	if client.Reconnects > 0 {

	}
}

func (client *Client) tryConnect() {
	var err error
	client.Connection, err = net.DialTimeout("tcp", client.Settings.Address, client.Settings.Timeout)

	if err != nil {
		time.Sleep(client.Settings.Timeout)
		return
	}

	//Got connection, initiate authentication
	client.rw = bufio.NewReadWriter(
		bufio.NewReader(client.Connection),
		bufio.NewWriter(client.Connection))

	//wait for authentication request
	auth := readMessage(client.rw)
	if auth.Type() != EventAuth {
		//invalid handshake
		return
	}

	//send password
	pwBuf := bytes.NewBufferString("auth ")
	pwBuf.WriteString(client.Settings.Password)
	pwBuf.WriteString("\n\n")
	client.rw.Write(pwBuf.Bytes())
	err = client.rw.Flush()

	if err != nil {
		client.Connection.Close()
		return
	}

	authResp := readMessage(client.rw)

	if authResp.Type() != EventReply {
		client.Connection.Close()
		return
	}

	if !authResp.Success() {
		client.Connection.Close()
		return
	}

	//connected and authed
	client.Connected = true
	client.Reconnects += 1
	if client.Settings.EventsChannel != nil {
		client.Settings.EventsChannel <- connectionState{connected: true}
	}
}

// Loop communicates with main process via Handler's message channel
// To be called after authentication or new connection
func (client *Client) Loop() {
	for {
		if !client.Connected {
			client.tryConnect()
			continue
		}

		message := readMessage(client.rw)

		switch message.Type() {
		case EventError:
			//disconnect
			client.Connection.Close()
			client.Connected = false
			if client.Settings.EventsChannel != nil {
				client.Settings.EventsChannel <- connectionState{connected: false}
			}
		case EventReply:
			client.executeChan <- message
		case EventGeneric:
			client.Settings.EventsChannel <- message
		case EventApi:
			client.apiChan <- message
		}

	}
}

func (client *Client) Api(cmd *Command) (Event, error) {
	client.lock.Lock()
	defer client.lock.Unlock()

	sendBytes(client.rw, cmd.GetSend())
	evt := <-client.apiChan

	return evt, nil
}

func (client *Client) Subscribe(evt string) (Event, error) {
	client.lock.Lock()
	defer client.lock.Unlock()
	cmd := Command{
		App:  "eventplain",
		Args: evt,
	}
	sendBytes(client.rw, cmd.GetSend())
	ret := <-client.executeChan

	return ret, nil
}

func (client *Client) Execute(cmd *Command) (Event, error) {
	client.lock.Lock()
	defer client.lock.Unlock()

	sendBytes(client.rw, cmd.GetExecute())
	evt := <-client.executeChan

	return evt, nil
}

// readMessage blocks until a message is available
func readMessage(rw *bufio.ReadWriter) Event {
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
	return new(eventReply)
}

func sendBytes(rw *bufio.ReadWriter, b []byte) (int, error) {
	defer rw.Flush()
	return rw.Write(b)
}
