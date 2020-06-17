package commands

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/tzvetkoff/dockerhub-watcher/pkg"
)

// NewPeriodicCommand ...
func NewPeriodicCommand() *cobra.Command {
	configPath := "./config.yml"

	cmd := &cobra.Command{
		Use:   "periodic",
		Short: "Start a periodic scanner",
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

			for {
				err = scanner.Scan()
				if err != nil {
					return err
				}

				time.Sleep(time.Duration(config.Periodic.Period) * time.Second)
			}
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", configPath, "Config")

	return cmd
}
