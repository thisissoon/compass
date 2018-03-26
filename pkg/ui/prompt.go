package ui

import (
	"gopkg.in/AlecAivazis/survey.v1/core"

	survey "gopkg.in/AlecAivazis/survey.v1"
)

// Module initialiser
func init() {
	core.QuestionIcon = ">"
}

// A ConfirmOption can customise a confirm message prompt
type ConfirmOption func(*survey.Confirm)

// ConfirmMessage returns a ConfirmOption to overrie the confirm prompt
func ConfirmMessage(v string) ConfirmOption {
	return func(p *survey.Confirm) {
		p.Message = v
	}
}

// ConfirmDefault returns a ConfirmOption that overrides the default value
// of a confirm prompt
func ConfirmDefault(v bool) ConfirmOption {
	return func(p *survey.Confirm) {
		p.Default = v
	}
}

// Confirm asks the user to enter yes or now returning a bool
func Confirm(opts ...ConfirmOption) bool {
	var v bool
	prompt := &survey.Confirm{
		Message: "Confirm?",
	}
	for _, opt := range opts {
		opt(prompt)
	}
	survey.AskOne(prompt, &v, nil)
	return v
}

// InputOption is a function that can customise an input option
type InputOption func(*survey.Input)

// InputMessage returns an InputOption that sets an input prompt message
func InputMessage(v string) InputOption {
	return func(p *survey.Input) {
		p.Message = v
	}
}

// InputDefault returns an InputOption that sets an input prompt default
func InputDefault(v string) InputOption {
	return func(p *survey.Input) {
		p.Default = v
	}
}

// Input prompts the user for input returning the input string
func Input(opts ...InputOption) string {
	var v string
	prompt := &survey.Input{
		Message: "Input:",
	}
	for _, opt := range opts {
		opt(prompt)
	}
	survey.AskOne(prompt, &v, nil)
	return v
}

// A ChoiceOption can customise a select prompt
type ChoiceOption func(*survey.Select)

// ChoiceMessage returns a ChoiceOption that sets the select prompt message
func ChoiceMessage(v string) ChoiceOption {
	return func(p *survey.Select) {
		p.Message = v
	}
}

// ChoiceDefault returns a ChoiceOption that sets the default select option
func ChoiceDefault(d string) ChoiceOption {
	return func(p *survey.Select) {
		p.Default = d
	}
}

// Choice returns the selected index of the provided choices
func Choice(choices []string, opts ...ChoiceOption) (int, bool) {
	prompt := &survey.Select{
		Message: "Select one of the following:",
		Options: choices,
	}
	for _, opt := range opts {
		opt(prompt)
	}
	var selected string
	survey.AskOne(prompt, &selected, nil)
	for i, choice := range choices {
		if selected == choice {
			return i, true
		}
	}
	return 0, false
}
