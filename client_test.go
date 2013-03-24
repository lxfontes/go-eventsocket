package eventsocket

import (
	"fmt"
	"testing"
	"time"
)

func Test_Client(t *testing.T) {
	t.Log("Test Basic Client")

	client, err := CreateClient(&ClientSettings{
		Address:  "localhost:8021",
		Password: "fongopass",
		Timeout:  10 * time.Second,
	})

	if err != nil {
		fmt.Println("Something went wrong ", err)
	}

	disconnecting := false
	for loop := true; loop; {
		msg := <-client.EventsChannel

		//For clients, 'client' and 'msg.Connection' will point to the same struct
		//However, 'client' has a few more fields available
		switch {
		case msg.Type == EventState && !msg.Success:
			fmt.Println("Disconnected")
			loop = false
		case msg.Type == EventState && msg.Success:
			fmt.Println("Connected")
			resp, err := msg.Connection.Send("event", "json", "RELOADXML")
			fmt.Println(resp, err)
			resp, err = msg.Connection.Send("api", "reloadxml")

			fmt.Println(resp, err)
		case msg.Type == EventGeneric:
			//get one event and disconnect
			if !disconnecting {
				msg.Connection.Send("exit")
				disconnecting = true
			}
		}
	}

}
