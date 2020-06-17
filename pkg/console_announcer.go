package pkg

import (
	"fmt"

	"github.com/fatih/color"
)

// ConsoleAnnouncer ...
type ConsoleAnnouncer struct {
	Color bool
}

// NewConsoleAnnouncer ...
func NewConsoleAnnouncer(config *ConsoleAnnouncerConfig) *ConsoleAnnouncer {
	return &ConsoleAnnouncer{
		Color: config.Color,
	}
}

// Announce ....
func (a *ConsoleAnnouncer) Announce(status string, repo string, tag *DockerTag) {
	output := fmt.Sprintf("%s repo=%s tag=%s", status, repo, tag.Name)

	if a.Color {
		switch status {
		case "=":
			output = color.WhiteString(output)
		case "+":
			output = color.GreenString(output)
		case "-":
			output = color.RedString(output)
		case "~":
			output = color.YellowString(output)
		}
	}

	fmt.Println(output)
}
