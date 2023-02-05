package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/robfig/cron/v3"
)

const peekLog = 1024

type Config struct {
	Schedule     string           `long:"schedule" env:"SCHEDULE" description:"Backup schedule" default:"@daily"`
	Dir          string           `long:"dir" env:"DIR" description:"Directory to backup and restore" default:"/data"`
	Prune        string           `long:"prune" env:"PRUNE" description:"Cleanup and prune backups schedule" default:"@daily"`
	Depth        int              `long:"depth" env:"DEPTH" description:"How many snapshots to keep" default:"7"`
	Restic       string           `long:"restic" env:"RESTIC" description:"Restic binary" default:"restic"`
	LogLimit     int              `long:"log-limit" env:"LOG_LIMIT" description:"Maximum number of bytes to be stored in memory from logs for notifications" default:"8192"`
	Notification HTTPNotification `group:"Webhook notification" namespace:"notification" env-namespace:"NOTIFICATION"`
}

type App struct {
	lock    sync.Mutex
	cfg     Config
	restic  binary
	running int32
	ready   int32
}

func New(cfg Config) *App {
	return &App{
		cfg:    cfg,
		restic: binary(cfg.Restic),
	}
}

func (app *App) Ready() bool {
	return atomic.LoadInt32(&app.ready) != 0
}

func (app *App) Run(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&app.running, 0, 1) {
		return fmt.Errorf("already running")
	}
	defer atomic.StoreInt32(&app.running, 0)

	err := app.runInit(ctx)
	if err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	if err := app.runRestore(ctx); err != nil {
		return fmt.Errorf("restore: %w", err)
	}

	crontab := cron.New()
	_, err = crontab.AddFunc(app.cfg.Schedule, func() {
		app.taskBackup(ctx)
	})
	if err != nil {
		return fmt.Errorf("add backup job: %w", err)
	}

	_, err = crontab.AddFunc(app.cfg.Prune, func() {
		app.taskPrune(ctx)
	})
	if err != nil {
		return fmt.Errorf("add prune job: %w", err)
	}

	atomic.StoreInt32(&app.ready, 1)
	defer atomic.StoreInt32(&app.ready, 0)
	log.Println("ready")
	crontab.Start()
	<-ctx.Done()
	sctx := crontab.Stop()
	<-sctx.Done()
	return sctx.Err()
}

func (app *App) taskPrune(ctx context.Context) {
	app.lock.Lock()
	defer app.lock.Unlock()
	app.cfg.Notification.Auto(ctx, "prune", func() error {
		return app.runPrune(ctx)
	})
}

func (app *App) taskBackup(ctx context.Context) {
	app.lock.Lock()
	defer app.lock.Unlock()
	app.cfg.Notification.Auto(ctx, "backup", func() error {
		return app.runBackup(ctx)
	})
}

func (app *App) runInit(ctx context.Context) error {
	var buffer = newLimitedBuffer(peekLog, os.Stderr)
	err := app.restic.invoke(ctx, buffer, "init")
	if err == nil || isAlreadyInitialized(buffer.buffer.String()) {
		return nil
	}
	return err
}

func (app *App) runBackup(ctx context.Context) error {
	var buffer = newLimitedBuffer(app.cfg.LogLimit, os.Stderr)

	if err := app.restic.invoke(ctx, buffer, "backup", app.cfg.Dir); err != nil {
		return fmt.Errorf("%s: %w", buffer.buffer.String(), err)
	}
	return nil
}

func (app *App) runRestore(ctx context.Context) error {
	markerFile := filepath.Join(app.cfg.Dir, ".restored")
	_, err := os.Stat(markerFile)
	if err == nil {
		log.Println("marker file exists - skipping restore")
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("check marker: %w", err)
	}

	var buffer = newLimitedBuffer(peekLog, os.Stderr)
	err = app.restic.invoke(ctx, buffer, "restore", "latest", "--target", "/")
	if err == nil || isNoSnapshot(buffer.buffer.String()) {
		return os.WriteFile(markerFile, []byte{}, 0755) // save marker
	}
	return err
}

func (app *App) runPrune(ctx context.Context) error {
	var buffer = newLimitedBuffer(app.cfg.LogLimit, os.Stderr)
	err := app.restic.invoke(ctx, buffer, "forget", "--prune", "--keep-last", strconv.Itoa(app.cfg.Depth))
	if err != nil {
		return fmt.Errorf("%s: %w", buffer.buffer.String(), err)
	}
	return nil
}
