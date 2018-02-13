package cmd

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	app := new(App)

	cmd := &cobra.Command{
		Use:   "exporter-merger",
		Short: "merges Prometheus metrics from multiple sources",
		Run:   app.run,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetLevel(log.DebugLevel)
		},
	}

	cmd.PersistentFlags().StringVarP(
		&app.configPath, "config-path", "c", "./merger.yaml",
		"Path to the configuration file.")
	cmd.PersistentFlags().IntVar(
		&app.port, "listen-port", 8080,
		"Listen port for the HTTP server.")

	cmd.AddCommand(NewVersionCommand())

	return cmd
}

type App struct {
	configPath string
	port       int
}

func (app *App) run(cmd *cobra.Command, args []string) {
	config, err := ReadConfig(app.configPath)
	if err != nil {
		log.WithField("error", err).Error("failed to load config")
		return
	}

	http.Handle("/metrics", Handler{
		Config: *config,
	})

	log.Infof("starting HTTP server on port %d", app.port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", app.port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
