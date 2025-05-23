package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func main() {
	fmt.Println("Starting Peril client...")
	const url = "amqp://guest:guest@localhost:5672"
	connection, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("could not connect to rabbitmq %v", err)
	}
	defer connection.Close()

	log.Println("connection established with rabbitmq client")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Printf("could not get username from stdin %v", err)
		return
	}

	_, queue, err := pubsub.DeclareAnBind(
		connection,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient)

	if err != nil {
		log.Fatalf("count not subscribe to %s", routing.PauseKey)
	}

	fmt.Printf("queue declared and bound: %s\n", queue.Name)

	gameState := gamelogic.NewGameState(username)

	for {
		commands := gamelogic.GetInput()

		if len(commands) == 0 {
			continue
		}

		action := commands[0]

		if action == "quit" {
			return
		}

		words := commands[:]

		err = runCommand(gameState, action, words)
		if err != nil {
			return
		}

	}

	//exit := make(chan os.Signal, 1)
	//
	//signal.Notify(exit, syscall.SIGINT)
	//
	//if _, ok := <-exit; ok {
	//	log.Printf("received signal to shut down client server")
	//	os.Exit(0)
	//}
}

func runCommand(gameState *gamelogic.GameState, action string, words []string) error {
	var err error
	switch action {
	case "spawn":
		err = gameState.CommandSpawn(words)
		if err != nil {
			log.Println(err)
		}
	case "move":
		armyMove, err := gameState.CommandMove(words)
		if err != nil {
			log.Println(err)
		}
		log.Printf("moved player: %s to location: %s\n", armyMove.Player.Username, armyMove.ToLocation)
	case "status":
		gameState.CommandStatus()
	case "help":
		gamelogic.PrintClientHelp()
	case "spam":
		fmt.Println("Spamming not allowed yet!")
	default:
		fmt.Println("Unknown command")
	}

	return err

}
