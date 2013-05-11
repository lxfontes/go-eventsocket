package eventsocket

import (
	"fmt"
	"testing"
)

type ServerNooper struct{}

func (n *ServerNooper) OnConnect(con *Connection) {
	fmt.Println("Connected")
	evt, err := con.Send("event", "plain", "HEARTBEAT")
	if err != nil {
		fmt.Println("ahn?")
		return
	}
	fmt.Println(evt.Success)

	con.Execute(&Command{
		App:  "answer",
		Args: "",
	})
}

func (n *ServerNooper) OnDisconnect(con *Connection) {
	fmt.Println("disconnect")
}

func (n *ServerNooper) OnEvent(con *Connection, evt *Event) {
	fmt.Println("event", evt.EventBody.Get("Event-Name"))
}

func (n *ServerNooper) OnClose(con *Connection) {
	fmt.Println("close")
}

func (n *ServerNooper) OnNewConnection(con *Connection) {
	fmt.Println("Got new connection")
	con.Listener = n
	go con.Loop()
}

func Test_Server(t *testing.T) {
	fmt.Println("Test Basic Server")

	server, err := CreateServer(ServerSettings{
		Address:  "localhost:8087",
		Listener: new(ServerNooper),
	})

	if err != nil {
		fmt.Println("Something went wrong ", err)
	}
	fmt.Println("Listening")

	server.Loop()

}
