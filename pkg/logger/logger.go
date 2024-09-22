package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
type Logger struct {
	*zap.SugaredLogger
}

// New returns a newly built Logger including all or no Options for configuration
func New(opts ...Option) *Logger {
	cfg := DefaultConfig
	for _, opt := range opts {
		opt(cfg)
	}

	return &Logger{
		zap.Must(cfg.Build()).Sugar(),
	}
}

// WithCustomConfig overrides the entire DefaultConfig configuration, replacing the reference with a new zap.Config
// object which will be used to configure the Logger
func WithCustomConfig(cfg *zap.Config) Option {
	return func(config *zap.Config) {
		config = cfg
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
