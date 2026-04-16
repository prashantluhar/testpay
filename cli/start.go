package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/prashantluhar/testpay/internal/api"
	"github.com/prashantluhar/testpay/internal/config"
	pgstore "github.com/prashantluhar/testpay/internal/store/postgres"
)

var configPath string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the TestPay mock server",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

		cfg, err := config.LoadFromFile(configPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		log.Info().Str("environment", cfg.Environment).Msg("config loaded")

		ctx := context.Background()
		pool, err := pgstore.ConnectPool(ctx, cfg.Database.URL)
		if err != nil {
			return fmt.Errorf("connecting to database: %w", err)
		}
		defer pool.Close()

		s := pgstore.New(pool)
		if err := pgstore.RunMigrations(pool); err != nil {
			return fmt.Errorf("running migrations: %w", err)
		}
		if err := pgstore.SeedLocalWorkspace(ctx, s); err != nil {
			return fmt.Errorf("seeding workspace: %w", err)
		}

		srv := api.NewServer(cfg, s)

		log.Info().Msgf("TestPay running on %s:%d (mode=%s)", cfg.Server.Host, cfg.Server.Port, cfg.Server.Mode)

		go func() {
			if err := srv.ListenAndServe(); err != nil {
				log.Error().Err(err).Msg("server error")
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration(cfg.Server.ShutdownTimeoutSeconds)*time.Second,
		)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	},
}

func init() {
	startCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to testpay.yaml config file")
}
