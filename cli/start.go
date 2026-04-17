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
	"github.com/prashantluhar/testpay/internal/trimmer"
	"github.com/prashantluhar/testpay/web"
)


var (
	configPath    string
	noDashboard   bool
	dashboardPort int
)

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
		if err := pgstore.SeedLocalWorkspace(ctx, s, cfg.Retention.DemoWorkspaceDailyCap); err != nil {
			return fmt.Errorf("seeding workspace: %w", err)
		}

		srv := api.NewServer(cfg, s)

		log.Info().Msgf("TestPay API running on %s:%d (mode=%s)", cfg.Server.Host, cfg.Server.Port, cfg.Server.Mode)

		// Background log trimmer — interval + retention window both come
		// from the loaded config (cfg.Retention.*), with 60-min / 10-day
		// defaults.
		trimmerCtx, cancelTrimmer := context.WithCancel(context.Background())
		defer cancelTrimmer()
		go trimmer.Run(
			trimmerCtx, s,
			time.Duration(cfg.Retention.TrimmerIntervalMinutes)*time.Minute,
			time.Duration(cfg.Retention.LogRetentionDays)*24*time.Hour,
		)

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Error().Err(err).Msg("server error")
			}
		}()

		// Optional embedded dashboard server.
		// Skipped in dev workflow where `pnpm dev` runs on the same port.
		var dash *http.Server
		if !noDashboard {
			dashFS, err := fs.Sub(web.Assets, "out")
			if err != nil {
				return fmt.Errorf("loading embedded dashboard: %w", err)
			}
			dash = &http.Server{
				Addr:    fmt.Sprintf(":%d", dashboardPort),
				Handler: http.FileServer(http.FS(dashFS)),
			}
			go func() {
				log.Info().Msgf("TestPay dashboard running at http://localhost:%d", dashboardPort)
				if err := dash.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error().Err(err).Msg("dashboard server error")
				}
			}()
		} else {
			log.Info().Msg("embedded dashboard disabled (--no-dashboard); run `pnpm dev` in web/ separately")
		}

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration(cfg.Server.ShutdownTimeoutSeconds)*time.Second,
		)
		defer cancel()
		if dash != nil {
			_ = dash.Shutdown(shutdownCtx)
		}
		return srv.Shutdown(shutdownCtx)
	},
}

func init() {
	startCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to testpay.yaml config file")
	startCmd.Flags().BoolVar(&noDashboard, "no-dashboard", false, "Disable the embedded dashboard (use `pnpm dev` in web/ instead)")
	startCmd.Flags().IntVar(&dashboardPort, "dashboard-port", 7701, "Port for the embedded dashboard")
}
