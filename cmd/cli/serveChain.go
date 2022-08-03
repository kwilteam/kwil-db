package cli

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/kwilteam/kwil-db/cmd/cli/commands"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:     "serve_chain",
	Aliases: []string{"test"},
	Short:   "Reverses a string",
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		var out []byte
		reset, err := cmd.Flags().GetBool("reset")
		if err != nil {
			log.Fatal(err)
		}
		if reset {
			fmt.Println("Reset called")
			bCmd := exec.Command("/bin/bash", "-c", "sh scripts/chain_serve_reset.sh")
			err = bCmd.Start()
			fmt.Println(err)
			//bCmd.
			err = commands.DeleteLogs()
			/*()
			fmt.Println("1")
			str, err := exec.LookPath("./")
			fmt.Println("str")
			fmt.Println(str)
			err = exec.Command("/bin/bash", "sh scripts/chain_serve_reset.sh").Run()
			fmt.Println(err)
			if err != nil {
				log.Fatal(err)
			}
			*/

		} else {
			out, err = exec.Command("/bin/bash", "-c", "sh scripts/chain_serve.sh").Output()
		}

		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(out))
		/*bCmd := exec.Command("/bin/sh", "chain_serve.sh")
		err := bCmd.Run()
		fmt.Println("hi")
		if err != nil {
			fmt.Println(err)
			log.Fatal(err)
		}
		//fmt.Println(out)*/
	},
}

var Source string

func init() {
	RootCmd.AddCommand(serveCmd)
	serveCmd.PersistentFlags().BoolP("reset", "r", false, "Reset the chain")
	//serveCmd.PersistentFlags().String("reset", "", "Reset the state of the chain and the logs")
	//serveCmd.Flags().StringVarP(&Source, "source", "s", "", "Source directory to read from")
}
