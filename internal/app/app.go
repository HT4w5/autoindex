package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/HT4w5/index/internal/config"
	"github.com/HT4w5/index/pkg/index"
	"github.com/HT4w5/index/pkg/log"
	"github.com/valyala/fasthttp"
)

type Application struct {
	cfg config.Config

	index   *index.Index
	httpsrv *fasthttp.Server
	logger  log.Logger
}

func New(cfg config.Config) *Application {
	return &Application{
		cfg: cfg,
	}
}

func (app *Application) Start() error {
	var level log.LogLevel
	switch strings.ToLower(app.cfg.Log.Level) {
	case "none":
		level = log.None
	case "error":
		level = log.Error
	case "warn":
		level = log.Warn
	case "":
		fallthrough
	case "info":
		level = log.Info
	case "debug":
		level = log.Debug
	}
	app.logger = &log.SimpleLogger{
		Level: level,
	}

	app.logger.Infof("starting application")

	// Create index
	var err error
	opts := make([]func(*index.Index), 0)
	if app.cfg.Filesystem.Root != "" {
		opts = append(opts, index.WithRoot(app.cfg.Filesystem.Root))
	}
	if app.cfg.Cache.TTL != 0 {
		opts = append(opts, index.WithTTL(time.Duration(app.cfg.Cache.TTL)))
	}
	if app.cfg.Cache.MaxSize != 0 {
		opts = append(opts, index.WithMaxSize(app.cfg.Cache.MaxSize))
	}

	opts = append(opts, index.WithLogger(app.logger))

	app.index, err = index.New(opts...)
	if err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}

	// HTTP listen
	app.httpsrv = &fasthttp.Server{
		Handler:      app.HandleQuery,
		IdleTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	addr := app.cfg.HTTP.Addr
	port := app.cfg.HTTP.Port
	if len(addr) == 0 {
		addr = "[::]"
	}
	if port == 0 {
		port = 80
	}

	go app.httpsrv.ListenAndServe(fmt.Sprintf("%s:%d", addr, port))

	app.logger.Infof("listening at http://%s:%d", addr, port)

	return nil
}

func (app *Application) Shutdown() error {
	app.logger.Infof("shutting down application")
	// HTTP shutdown
	err := app.httpsrv.Shutdown()

	// Index close
	return errors.Join(err, app.index.Close())
}
