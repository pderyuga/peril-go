package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	rabbitConnString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")

	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not open channel: %v", err)
	}

	_, queue, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, fmt.Sprintf("%s.*", routing.GameLogSlug), pubsub.QueueTypeDurable)
	if err != nil {
		log.Fatalf("could not subscribe to game logs: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	gamelogic.PrintClientHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "pause":
			fmt.Println("Publishing paused game state..")
			jsonData, err := json.Marshal(routing.PlayingState{IsPaused: true})

			err = pubsub.PublishJSON(publishCh, routing.ExchangePerilDirect, routing.PauseKey, jsonData)
			if err != nil {
				log.Printf("could not publish: %v", err)
			}
			fmt.Println("Pause message sent!")
		case "resume":
			fmt.Println("Publishing resume game state...")
			jsonData, err := json.Marshal(routing.PlayingState{IsPaused: false})

			err = pubsub.PublishJSON(publishCh, routing.ExchangePerilDirect, routing.PauseKey, jsonData)
			if err != nil {
				log.Printf("could not publish: %v", err)
			}
			fmt.Println("Resume message sent!")
		case "quit":
			fmt.Println("Exiting game...")
			os.Exit(0)
		default:
			fmt.Println("Command not recognized")
			continue
		}
	}
}
