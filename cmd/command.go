package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exporter-merger",
		Short: "merges Prometheus metrics from multiple sources",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetLevel(log.DebugLevel)
		},
	}

	cmd.AddCommand(NewVersionCommand())

	return cmd
}
