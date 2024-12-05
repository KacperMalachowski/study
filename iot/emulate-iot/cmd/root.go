package cmd

import "github.com/spf13/cobra"

var (
	// Used for flags.
	RabbitMQHost string
	RabbitMQPort string
	RabbitMQUser string
	RabbitMQPass string
)

var rootCmd = &cobra.Command{
	Use:   "iot",
	Short: "iot is a CLI for emulating IoT devices",
	Long:  `iot is a CLI for emulating IoT devices`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&RabbitMQHost, "rabbitmq-host", "167.235.62.0", "RabbitMQ host")
	rootCmd.PersistentFlags().StringVar(&RabbitMQPort, "rabbitmq-port", "5672", "RabbitMQ port")
	rootCmd.PersistentFlags().StringVar(&RabbitMQUser, "rabbitmq-user", "emulator", "RabbitMQ user")
	rootCmd.PersistentFlags().StringVar(&RabbitMQPass, "rabbitmq-pass", "emulator", "RabbitMQ password")
}

func Execute() {
	rootCmd.Execute()
}
