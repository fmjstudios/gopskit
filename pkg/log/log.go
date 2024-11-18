package log

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"os"
	"sync"
	"time"
)

var (
	// DefaultConfig is the default configuration used for the zap-based Logger
	DefaultConfig = &zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "msg",
			LevelKey:    "lvl",
			EncodeLevel: zapcore.CapitalColorLevelEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: nil,
		InitialFields: map[string]interface{}{
			"date": time.Now().Format(time.RFC3339),
		},
	}
)

// Option is a utility function which is called during Logger initialization to alter or even override the DefaultConfig
type Option func(config *zap.Config)

// Logger is a type alias to Uber's zap logger, saving us from importing zap everywhere
// and assuming the SugaredLogger's API results low effort logs later
type Logger struct {
	// log is the underlying zap Logger which is later embedded into
	// our Logger
	log *zap.Logger

	// conf is private zap.Config for the underlying instance within log
	conf *zap.Config

	// lock is a Mutex which ensures that only one goroutine may modify the configuration
	lock sync.Mutex

	// Logger embeds a pointer zap.SugaredLogger to assume its' API
	*zap.SugaredLogger
}

// Config builds a new zap.Config with optional Options for configuration
func Config(opts ...zap.Option) *zap.Config {
	c := zap.NewProductionConfig()
	g := new(errgroup.Group)

	// assert that it builds
	g.Go(func() error {
		_, err := c.Build(opts...)
		if err != nil {
			return err
		}

		return nil
	})

	// crash if it isn't so
	if err := g.Wait(); err != nil {
		fmt.Printf("could not create zap.Config for Logger: %v", err)
		os.Exit(1)
	}

	return &c
}

// New returns a newly built Logger including all or no Options for configuration
func New(opts ...Option) *Logger {
	l := &Logger{
		conf: Config(),
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	errg, _ := errgroup.WithContext(context.Background())
	for _, opt := range opts {
		errg.Go(func() error {
			opt(l.conf)
			return nil
		})
	}

	// crash if it isn't so
	if err := errg.Wait(); err != nil {
		fmt.Printf("could not configure zap.Config for Logger: %v", err)
		os.Exit(1)
	}

	lgr, err := l.conf.Build()
	if err != nil {
		fmt.Printf("could not build Config for Logger: %v", err)
	}

	return &Logger{
		log:           lgr,
		SugaredLogger: lgr.Sugar(),
	}
}

// WithCustomConfig overrides the entire DefaultConfig configuration, replacing the reference with a new zap.Config
// object which will be used to configure the Logger
func WithCustomConfig(cfg zap.Config) Option {
	return func(config *zap.Config) {
		config = &cfg
	}
}

// WithLevel configures the Logger with a custom logging level
func WithLevel(level zapcore.Level) Option {
	return func(config *zap.Config) {
		config.Level = zap.NewAtomicLevelAt(level)
	}
}

// WithEncoder configures a custom Encoder for the new Logger
func WithEncoder(encoder zapcore.EncoderConfig) Option {
	return func(config *zap.Config) {
		config.EncoderConfig = encoder
	}
}

// WithDevelopment configures the new Logger for use within development contexts
func WithDevelopment() Option {
	return func(config *zap.Config) {
		config.Development = true
	}
}
