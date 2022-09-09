package queries

import (
	"fmt"
	"github.com/kwilteam/kwil-db/cli/cmd/utils"
)

func Queries() {
	input, err := utils.PromptStringArr("Please choose what action you would like to perform", []string{"Create", "Delete"})
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	switch input {
	case "create":
		CreateQuery()
	case "delete":
		DeleteQuery()
	}
}
