package main

import (
	"log"

	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/cmd/kwild/internal/util"
)

func main() {
	err := util.BuildAndRunRootCommand()
	if err != nil {
		log.Fatalf("Error while running kwild: %v", err)
	}
}
