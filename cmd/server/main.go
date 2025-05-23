package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"slices"
)

func main() {
	fmt.Println("Starting Peril server...")
	url := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("could not connect to rabbitmq server %v", err)
	}
	defer connection.Close()

	fmt.Println("connection to rabbitmq established successfully")

	channel, err := connection.Channel()
	if err != nil {
		log.Fatalf("could not create a channel on the rabbitmq server %v", err)
	}
	defer channel.Close()

	_, logQueue, err := pubsub.DeclareAnBind(
		connection,
		routing.ExchangePerilTopic,
		"game_logs",
		routing.GameLogSlug,
		pubsub.Durable,
	)
	if err != nil {
		log.Fatalf("could not declare and bind queue: %v", err)
	}

	log.Printf("Declare and bind queue successful: %s\n", logQueue.Name)

	gamelogic.PrintServerHelp()

	for {
		userInput := gamelogic.GetInput()

		if len(userInput) == 0 {
			continue
		}

		action := userInput[0]

		if action == "quit" {
			break
		}

		err = sendMessage(action, channel)

		if err != nil {
			log.Fatalf("could not publish message to rabbitmq %v", err)
		}
	}

}

func sendMessage(action string, channel *amqp.Channel) error {
	var playingState routing.PlayingState

	switch action {
	case "pause":
		playingState.IsPaused = true
		logAction(action)
	case "resume":
		playingState.IsPaused = false
		logAction(action)
	default:
		logAction(action, "unknown")
		return nil
	}

	err := pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, playingState)
	if err != nil {
		return fmt.Errorf("error publishing message to exchange %v", err)
	}

	return err
}

func logAction(action string, args ...string) {
	if slices.Contains(args, "unknown") {
		log.Println("Unknown action")
		return
	}
	log.Printf("sending %s action\n", action)
}
