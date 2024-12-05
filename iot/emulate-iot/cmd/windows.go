package cmd

import (
	"log"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/cobra"
)

var windowsCmd = &cobra.Command{
	Use:   "windows",
	Short: "Emulate a window controller",
	Long:  `Emulate a window controller that opens and closes windows based on a rabbitmq message`,
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := amqp.Dial("amqp://" + RabbitMQUser + ":" + RabbitMQPass + "@" + RabbitMQHost + ":" + RabbitMQPort)
		failOnError(err, "Failed to connect to RabbitMQ")
		defer conn.Close()

		ch, err := conn.Channel()
		failOnError(err, "Failed to open a channel")
		defer ch.Close()

		q, err := ch.QueueDeclare(
			"windows",
			true,
			true,
			false,
			false,
			nil,
		)
		failOnError(err, "Failed to declare a queue")

		err = ch.QueueBind(
			q.Name,
			"windows",
			"amq.topic",
			false,
			nil,
		)
		failOnError(err, "Failed to bind a queue")

		msgs, err := ch.Consume(
			q.Name,
			"",
			true,
			false,
			false,
			false,
			nil,
		)
		failOnError(err, "Failed to register a consumer")

		forever := make(chan bool)

		go func() {
			for d := range msgs {
				msg := strings.Trim(string(d.Body), "\"")
				log.Printf("Received a message: %s", msg)
				if msg == "open" {
					log.Println("Opening windows")
				} else if msg == "close" {
					log.Println("Closing windows")
				}
			}
		}()

		log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
		<-forever
	},
}

func init() {
	rootCmd.AddCommand(windowsCmd)
}
