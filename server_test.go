package eventsocket

import (
	"fmt"
	"testing"
)

func Test_Server(t *testing.T) {
	t.Log("Test Basic Server")

	server, err := CreateServer(&ServerSettings{
		Address: "localhost:8086",
	})

	if err != nil {
		fmt.Println("Something went wrong ", err)
	}

	fmt.Println("Listening on", server.Settings.Address)

	for loop := true; loop; {
		msg := <-server.EventsChannel
		switch {
		case msg.Type == EventState && !msg.Success:
			fmt.Println("Disconnected")
			loop = false
		case msg.Type == EventState && msg.Success:
			fmt.Println("Connected")
			fmt.Println("Answering")
			msg.Connection.Answer()

		case msg.Type == EventGeneric:

		}
	}

}
