package main

import (
	"github.com/spf13/cobra"
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.InfoLevel)

	log.SetFormatter(&log.JSONFormatter{})

	err := os.Setenv("TZ", "") // Use UTC by default :)
	if err != nil {
		log.Fatal("failed to set env - ", err)
	}

	rootcmd := &cobra.Command{
		Use:     "Emit",
		Version: "v1",
		Short:   "Emit events to mimic popular webhook providers",
	}

	parsePersistentArgs(rootcmd)
	if err := rootcmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func parsePersistentArgs(cmd *cobra.Command) {
	var url string
	var secret string
	cmd.PersistentFlags().StringVar(&url, "url", "", "URL to emit event to")
	cmd.PersistentFlags().StringVar(&secret, "secret", "", "secret to use to encode hmac header")

	cmd.AddCommand(addShopifyCommand())
	cmd.AddCommand(addGithubCommand())
}
