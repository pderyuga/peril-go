package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")

	rabbitConnString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game client connected to RabbitMQ!")

	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not get username: %v", err)
	}

	gs := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, username), routing.PauseKey, pubsub.QueueTypeTransient, handlerPause(gs))
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	err = pubsub.SubscribeJSON(
		conn,                       // connection
		routing.ExchangePerilTopic, // exchange
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username), // queue name
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),            // routing key
		pubsub.QueueTypeTransient,                               // queue type "transient" or "durable"
		handlerMove(gs))
	if err != nil {
		log.Fatalf("could not subscribe to move: %v", err)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			err := gs.CommandSpawn(words)
			if err != nil {
				fmt.Println("spawn command failed:", err)
				continue
			}
		case "move":
			move, err := gs.CommandMove(words)
			if err != nil {
				fmt.Println("move command failed:", err)
				continue
			}
			err = pubsub.PublishJSON(
				publishCh,                  //channel
				routing.ExchangePerilTopic, // exchange
				fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username), // key
				move, // json
			)
			if err != nil {
				fmt.Println("failed to publish move:", err)
				continue
			}
			fmt.Printf("Moved %v units to %s\n", len(move.Units), move.ToLocation)
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			os.Exit(0)
		default:
			fmt.Println("Command not recognized")
		}
	}
}
