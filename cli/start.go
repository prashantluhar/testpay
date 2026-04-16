package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/prashantluhar/testpay/internal/api"
	"github.com/prashantluhar/testpay/internal/config"
	pgstore "github.com/prashantluhar/testpay/internal/store/postgres"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the TestPay mock server",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		ctx := context.Background()
		pool, err := pgstore.ConnectPool(ctx, cfg.DatabaseURL)
		if err != nil {
			return fmt.Errorf("connecting to database: %w", err)
		}
		defer pool.Close()

		s := pgstore.New(pool)
		if err := pgstore.RunMigrations(pool); err != nil {
			return fmt.Errorf("running migrations: %w", err)
		}
		if err := pgstore.SeedLocalWorkspace(ctx, s); err != nil {
			return fmt.Errorf("seeding local workspace: %w", err)
		}

		srv := api.NewServer(cfg, s)

		log.Info().Msgf("TestPay mock server running at http://localhost:%d", cfg.Port)
		log.Info().Msg("Point your app at /stripe, /razorpay, or /v1")

		go func() {
			if err := srv.ListenAndServe(); err != nil {
				log.Error().Err(err).Msg("server error")
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Info().Msg("Shutting down...")
		return srv.Shutdown(context.Background())
	},
}
