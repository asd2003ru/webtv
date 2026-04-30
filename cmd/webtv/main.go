package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"webtv/internal/api"
	"webtv/internal/config"
	"webtv/internal/logx"
	"webtv/internal/scheduler"
	"webtv/internal/storage"
	"webtv/internal/stream"
	"webtv/internal/webui"
)

func main() {
	cfg := config.Load()
	level := logx.SetLevelFromString(cfg.LogLevel)
	logx.Infof("log level=%s", logx.LevelString(level))

	store, err := storage.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("storage init: %v", err)
	}
	defer store.Close()

	syncer := scheduler.NewSyncService(store)
	streamMgr := stream.NewManager(
		cfg.SingleConnPolicy,
		cfg.FFmpegBin,
		time.Duration(cfg.StreamIdleTimeoutSec)*time.Second,
		cfg.StreamRetryMax,
		cfg.ManifestBackoffEnable,
	)

	static := webui.Handler()
	srv := api.NewWithAppTitle(
		store,
		syncer,
		streamMgr,
		static,
		time.Duration(cfg.StreamModeCacheTTLMin)*time.Minute,
		cfg.AppTitle,
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	sched := scheduler.New(syncer, store)
	go sched.Start(ctx)

	httpSrv := &http.Server{Addr: cfg.Addr, Handler: srv.Handler()}
	go func() {
		logx.Infof("webtv listening on %s", cfg.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}
