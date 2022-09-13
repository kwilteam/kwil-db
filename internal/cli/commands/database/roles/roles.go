package roles

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/cli/utils"
)

func Roles() {
	input, err := utils.PromptStringArr("Please choose what action you would like to perform", []string{"Create", "Modify", "Delete"})
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	switch input {
	case "create":
		CreateRole()
	case "modify":
		ModifyRole()
	case "delete":
		DeleteRole()
	}
}
