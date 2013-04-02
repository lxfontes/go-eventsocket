package main

import (
	"flag"
	"fmt"
	esl "github.com/lxfontes/go-eventsocket"
	"time"
)

var address *string = flag.String("address", "localhost:8021", "Freeswitch Event Socket Address")
var password *string = flag.String("password", "cluecon", "Event Socket Password")

func main() {

	flag.Parse()

	fmt.Println("Connecting to", *address, "using password", *password)
	client, err := esl.CreateClient(&esl.ClientSettings{
		Address:  *address,
		Password: *password,
		Timeout:  10 * time.Second,
	})

	if err != nil {
		fmt.Println("Something went wrong ", err)
	}

	for loop := true; loop; {
		msg := <-client.EventsChannel

		switch {
		case msg.Type == esl.EventState && !msg.Success:
			fmt.Println("Disconnected")
			loop = false
		case msg.Type == esl.EventState && msg.Success:
			fmt.Println("Connected")
			_, err := client.Send("event", "plain", "ALL")
			if err != nil {
				fmt.Println("Something went wrong while subscribing to all events")
			}
		case msg.Type == esl.EventGeneric:
			//just print all events
			s := string(msg.Body)
			fmt.Println(s)
		}
	}

}
