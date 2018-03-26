package ui

import (
	"k8s.io/apimachinery/pkg/labels"
)

var asciiLogo = "" +
	"  _________  ____ ___  ____  ____ ___________\n" +
	" / ___/ __ \\/ __ `__ \\/ __ \\/ __ `/ ___/ ___/\n" +
	"/ /__/ /_/ / / / / / / /_/ / /_/ (__  |__  )\n" +
	"\\___/\\____/_/ /_/ /_/ .___/\\__,_/____/____/\n" +
	"                   /_/"

// LabelsPrompt prompts the user to enter kbernetes labels
func LabelsPrompt() labels.Set {
	labels := labels.Set{}
	for {
		name := Input(InputMessage("Name:"))
		value := Input(InputMessage("Value:"))
		labels[name] = value
		another := Confirm(
			ConfirmMessage("Add another label?"),
			ConfirmDefault(false))
		if !another {
			break
		}
	}
	return labels
}
