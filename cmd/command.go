package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRootCommand() *cobra.Command {
	app := new(App)

	cmd := &cobra.Command{
		Use:   "exporter-merger",
		Short: "merges Prometheus metrics from multiple sources",
		Run:   app.run}
	// PersistentPreRun: func(cmd *cobra.Command, args []string) {
	// 	log.SetLevel(log.DebugLevel)
	// },
	// }

	app.Bind(cmd)

	cmd.AddCommand(NewVersionCommand())

	return cmd
}

type App struct {
	viper *viper.Viper
}

func (app *App) Bind(cmd *cobra.Command) {
	app.viper = viper.New()
	app.viper.SetEnvPrefix("MERGER")
	app.viper.AutomaticEnv()

	configPath := cmd.PersistentFlags().StringP(
		"config-path", "c", "",
		"Path to the configuration file.")
	cobra.OnInitialize(func() {
		if configPath != nil && *configPath != "" {
			config, err := ReadConfig(*configPath)
			if err != nil {
				log.WithField("error", err).Errorf("failed to load config file '%s'", *configPath)
				os.Exit(1)
				return
			}

			urls := []string{}
			for _, e := range config.Exporters {
				urls = append(urls, e.URL)
			}
			app.viper.SetDefault("urls", strings.Join(urls, " "))
		}
	})

	cmd.PersistentFlags().Int(
		"listen-port", 8080,
		"Listen port for the HTTP server. (ENV:MERGER_PORT)")
	app.viper.BindPFlag("port", cmd.PersistentFlags().Lookup("listen-port"))

	cmd.PersistentFlags().Int(
		"exporters-timeout", 10,
		"HTTP client timeout for connecting to exporters. (ENV:MERGER_EXPORTERS_TIMEOUT)")
	app.viper.BindPFlag("exporters-timeout", cmd.PersistentFlags().Lookup("exporters-timeout"))

	cmd.PersistentFlags().String(
		"log-level", "error",
		"Log level (possible values: debug, info, warning, error, fatal, panic). (ENV:MERGER_LOG_LEVEL)")

	app.viper.BindPFlag("log-level", cmd.PersistentFlags().Lookup("log-level"))

	loglevel, err := log.ParseLevel(app.viper.GetString("log-level"))
	if err != nil {
		log.Fatalf("Error parsing log level: %v", err)
	}

	log.SetLevel(loglevel)

	cmd.PersistentFlags().StringSlice(
		"url", nil,
		"URL to scrape. Can be speficied multiple times. (ENV:MERGER_URLS,space-seperated)")
	app.viper.BindPFlag("urls", cmd.PersistentFlags().Lookup("url"))
}

func (app *App) run(cmd *cobra.Command, args []string) {
	http.Handle("/metrics", Handler{
		Exporters:            app.viper.GetStringSlice("urls"),
		ExportersHTTPTimeout: app.viper.GetInt("exporters-timeout"),
	})

	port := app.viper.GetInt("port")
	log.Infof("starting HTTP server on port %d", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
