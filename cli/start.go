package cli

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/prashantluhar/testpay/internal/api"
	"github.com/prashantluhar/testpay/internal/config"
	"github.com/prashantluhar/testpay/internal/observability"
	pgstore "github.com/prashantluhar/testpay/internal/store/postgres"
	"github.com/prashantluhar/testpay/web"
)

var configPath string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the TestPay mock server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadFromFile(configPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		if cfg.Server.Mode == "hosted" && cfg.Auth.JWTSecret == "" {
			return fmt.Errorf("JWT_SECRET is required when mode=hosted (set via env var referenced by auth.jwt_secret_env)")
		}
		observability.Setup(cfg.Logging.Level, cfg.Logging.Format, cfg.Environment, "testpay")
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

		log.Info().Msgf("TestPay API running on %s:%d (mode=%s)", cfg.Server.Host, cfg.Server.Port, cfg.Server.Mode)

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Error().Err(err).Msg("server error")
			}
		}()

		// Dashboard server on :7701 (embedded Next.js static export)
		dashFS, err := fs.Sub(web.Assets, "out")
		if err != nil {
			return fmt.Errorf("loading embedded dashboard: %w", err)
		}
		dash := &http.Server{
			Addr:    ":7701",
			Handler: http.FileServer(http.FS(dashFS)),
		}
		go func() {
			log.Info().Msg("TestPay dashboard running at http://localhost:7701")
			if err := dash.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Error().Err(err).Msg("dashboard server error")
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
		_ = dash.Shutdown(shutdownCtx)
		return srv.Shutdown(shutdownCtx)
	},
}

func init() {
	startCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to testpay.yaml config file")
}
