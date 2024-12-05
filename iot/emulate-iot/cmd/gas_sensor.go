package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/cobra"
)

var (
	gasLevel int  = 0
	enforce  bool = false
)

var gasSensorCmd = &cobra.Command{
	Use:   "gas",
	Short: "Emulate a gas sensor",
	Long:  `Emulate a gas sensor that sends random values to a RabbitMQ server`,
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := amqp.Dial("amqp://" + RabbitMQUser + ":" + RabbitMQPass + "@" + RabbitMQHost + ":" + RabbitMQPort)
		failOnError(err, "Failed to connect to RabbitMQ")
		defer conn.Close()

		ch, err := conn.Channel()
		failOnError(err, "Failed to open a channel")
		defer ch.Close()

		q, err := ch.QueueDeclare(
			"gas",
			true,
			true,
			false,
			false,
			nil,
		)
		failOnError(err, "Failed to declare a queue")

		alarmQ, err := ch.QueueDeclare(
			"alarm",
			false,
			true,
			false,
			false,
			nil,
		)
		failOnError(err, "Failed to declare a queue")

		windwsQ, err := ch.QueueDeclare(
			"windows",
			true,
			true,
			false,
			false,
			nil,
		)
		failOnError(err, "Failed to declare a queue")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for {
			log.Println("Gas sensor is running")
			if !enforce {
				randomGasLevelChange()
			}
			body := fmt.Sprintf("%d", gasLevel)
			err = ch.PublishWithContext(
				ctx,
				"amq.topic",
				q.Name,
				false,
				false,
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(body),
				},
			)
			failOnError(err, "Failed to publish a message")
			log.Printf(" [x] Sent %s", body)
			time.Sleep(1 * time.Second)

			if gasLevel > 15 {
				err = ch.PublishWithContext(
					ctx,
					"amq.topic",
					alarmQ.Name,
					false,
					false,
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte("true"),
					},
				)
				failOnError(err, "Failed to publish a message")
				log.Printf(" [x] Sent alarm")

				err = ch.PublishWithContext(
					ctx,
					"amq.topic",
					windwsQ.Name,
					false,
					false,
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte("open"),
					},
				)
				failOnError(err, "Failed to publish a message")
				log.Printf(" [x] Sent open windows")
			}
		}
	},
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func randomGasLevelChange() {
	decrease := randomBool()
	if decrease && gasLevel > 0 {
		gasLevel -= 1
	} else if !decrease && gasLevel < 30 {
		gasLevel += 1
	}
}

func randomBool() bool {
	return time.Now().UnixNano()%2 == 0
}

func init() {
	gasSensorCmd.Flags().IntVarP(&gasLevel, "level", "l", 0, "Initial gas level")
	gasSensorCmd.Flags().BoolVarP(&enforce, "enforce", "e", false, "Enforce gas level")

	rootCmd.AddCommand(gasSensorCmd)
}
