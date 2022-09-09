package table

import (
	"fmt"
	"github.com/kwilteam/kwil-db/cli/cmd/utils"
)

func Table() {
	input, err := utils.PromptStringArr("Please choose what action you would like to perform", []string{"Create", "Modify", "Delete"})
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	switch input {
	case "create":
		CreateTable()
	case "modify":
		ModifyTable()
	case "delete":
		DeleteTable()
	}
}
