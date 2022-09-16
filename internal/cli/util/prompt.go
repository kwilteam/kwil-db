package util

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

type Prompter struct {
	Label       string
	Default     string
	Validate    promptui.ValidateFunc
	MaskDefault bool
	ShowLast    int
	HideEntered bool
}

func (p Prompter) Run() (string, error) {
	defaultVal := p.Default
	if len(defaultVal) > 0 && p.MaskDefault {
		defaultVal = strings.Repeat("*", len(defaultVal)-p.ShowLast) + defaultVal[len(defaultVal)-p.ShowLast:]
	}

	valid := "{{ . | green }}"
	invalid := "{{ . | red }}"

	if len(defaultVal) > 0 {
		valid += fmt.Sprintf(` {{ "[%s]" | faint }}`, defaultVal)
		invalid += fmt.Sprintf(` {{ "[%s]" | faint }}`, defaultVal)
	}

	valid += ": "
	invalid += ": "

	prompt := promptui.Prompt{
		Label: p.Label,
		Templates: &promptui.PromptTemplates{
			Prompt:  "{{ . }} ",
			Valid:   valid,
			Invalid: invalid,
			Success: `{{ "✔" | green }} {{ . | bold }} `,
		},
		Validate:    p.Validate,
		Default:     p.Default,
		HideEntered: p.HideEntered,
		AllowEdit:   true,
	}

	return prompt.Run()
}
