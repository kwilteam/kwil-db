package utils

import (
	"github.com/manifoldco/promptui"
	"strings"
)

// promptStringArr prompts the given label with a string array and returns the lowercase result
func PromptStringArr(l string, items []string) (string, error) {
	prompt := promptui.Select{
		Label: l,
		Items: items,
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return strings.ToLower(result), nil
}

func PromptStringInput(l string) (string, error) {
	prompt := promptui.Prompt{
		Label: l,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}
