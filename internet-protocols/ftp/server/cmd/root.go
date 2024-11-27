package cmd

import (
	"log"

	"github.com/KacperMalachowski/study/internet-protocols/ftp/server/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile       string
	configuration = config.Config{}
)

var rootCmd = &cobra.Command{
	Use:   "usftp",
	Short: "usftp is a simple FTP server",
	Long:  "usftp is a simple FTP server that supports the following commands: USER, PASS, PWD, CDUP, CWD, MKD, RWD, LIST, STOR, RETR, DELE",
	Run: func(cmd *cobra.Command, args []string) {
		if err := configuration.Validate(); err != nil {
			log.Fatalf("Error validating configuration: %s", err)
		}
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
	rootCmd.PersistentFlags().VarP(&configuration.Users, "user", "u", "user info in the format of 'username:password:home_dir'")
	rootCmd.PersistentFlags().StringVarP(&configuration.Address, "address", "a", "0.0.0.0", "address on which the server will listen")
	rootCmd.PersistentFlags().IntVarP(&configuration.Port, "port", "p", 21, "port on which the server will listen")
	rootCmd.PersistentFlags().BoolVarP(&configuration.AllowAnonymous, "allow-anonymous", "A", false, "allow anonymous login")
	rootCmd.PersistentFlags().StringVarP(&configuration.RootDir, "root-dir", "r", "/tmp", "root directory of the server")
	rootCmd.PersistentFlags().IntVarP(&configuration.MinPassivePort, "min-passive-port", "m", 30000, "minimum port for passive mode")
	rootCmd.PersistentFlags().IntVarP(&configuration.MaxPassivePort, "max-passive-port", "M", 30010, "maximum port for passive mode")
	viper.BindPFlag("users", rootCmd.PersistentFlags().Lookup("user"))
	viper.BindPFlag("address", rootCmd.PersistentFlags().Lookup("address"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("allow_anonymous", rootCmd.PersistentFlags().Lookup("allow-anonymous"))
	viper.BindPFlag("root_dir", rootCmd.PersistentFlags().Lookup("root-dir"))
	viper.BindPFlag("min_passive_port", rootCmd.PersistentFlags().Lookup("min-passive-port"))
	viper.BindPFlag("max_passive_port", rootCmd.PersistentFlags().Lookup("max-passive-port"))
}

func initConfig() {
	viper.SetConfigType("yaml")

	if cfgFile == "" {
		return
	}

	log.Println("Reading config file...")
	viper.SetConfigFile(cfgFile)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config: %s", err)
	}

	if err := viper.Unmarshal(&configuration, viper.DecodeHook(config.UserStringDecodeHook())); err != nil {
		log.Fatalf("Error unmarshalling config: %s", err)
	}
	log.Println("Config file read successfully")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error: %s", err)
	}
}
