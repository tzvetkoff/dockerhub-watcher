package commands

import (
	"github.com/spf13/cobra"

	"github.com/tzvetkoff/dockerhub-watcher/pkg"
)

// NewRunCommand ...
func NewRunCommand() *cobra.Command {
	configPath := "./config.yml"

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Do a 1-time scan",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := pkg.LoadConfig(configPath)
			if err != nil {
				return err
			}

			dockerClient := pkg.NewDockerClient(&config.Docker)

			announcers := []pkg.Announcer{}
			if config.Console.Enabled {
				announcers = append(announcers, pkg.NewConsoleAnnouncer(&config.Console))
			}
			if config.Slack.Enabled {
				announcers = append(announcers, pkg.NewSlackAnnouncer(&config.Slack))
			}

			scanner := pkg.NewScanner(&config.Docker, dockerClient, announcers)

			return scanner.Scan()
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", configPath, "Config")

	return cmd
}
