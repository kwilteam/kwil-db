/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/fatih/color"
	"github.com/kwilteam/kwil-db/internal/cli/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect tests a connection with the Kwil node.",
	Long: `Connect tests a connection with the specified Kwil node.  It also
	exchanges relevant information regarding the node's capabilities, keys, etc..`,
	Run: func(cmd *cobra.Command, args []string) {
		err := utils.LoadConfig()
		if err != nil {
			c := color.New(color.FgRed)
			c.Println(err)
			return
		}

		// now we get the endpoint from viper
		endpoint := viper.GetString("endpoint")
		if endpoint == "" {
			c := color.New(color.FgRed)
			c.Println("Endpoint not set.  Please set an endpoint with 'kwil set endpoint <endpoint>'.")
			return
		}

		// we also need the JWT token
		token := viper.GetString("api-key")

		// now we need to send a get request to the endpoint /api/v0/connect
		// we need to send the token in the header
		req, err := http.NewRequest("GET", "http://"+endpoint+"/api/v0/connect", nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		// set the header
		req.Header.Set("Authorization", "Bearer "+token)

		// send the request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c := color.New(color.FgRed)
			c.Println("Error: ", err)
			return
		}

		// check the status code
		if resp.StatusCode != 200 {
			c := color.New(color.FgRed)
			c.Println("Error: " + resp.Status)
			return
		}

		// now we need to read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		// set the node-address to the body
		viper.Set("node-address", string(body))
		viper.WriteConfig()

		c := color.New(color.FgGreen)
		c.Println("connection successful")
	},
}

func init() {
	RootCmd.AddCommand(connectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// connectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// connectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
