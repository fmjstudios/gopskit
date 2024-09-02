package app

import (
	"github.com/fmjstudios/gopskit/pkg/cmd"
	"github.com/fmjstudios/gopskit/pkg/common"
	"github.com/fmjstudios/gopskit/pkg/kube"
	"github.com/fmjstudios/gopskit/pkg/logger"
	"github.com/fmjstudios/gopskit/pkg/platform"
	"github.com/fmjstudios/gopskit/pkg/stamp"
	"go.uber.org/zap"
)

const APP_NAME = "ssolo"

type AppOpt func(a *App)

// type App is the implementation for the `ssolo` command-line application
type App struct {
	*common.GOpsKitApp

	// KeyClient is a Keycloak HTTP client, which ssolo uses to configure
	// Keycloak to authenticate users for certain applications
}

// New creates a newly initialized instance of the App type
func New(opts ...AppOpt) *App {
	platform := platform.New(platform.WithApp(APP_NAME))
	logger := logger.New()
	defer logger.Sync()

	exec := cmd.NewExecutor(cmd.WithInheritedEnv())
	kc, err := kube.NewClient()
	if err != nil {
		logger.Fatal("could not create Kubernetes Client", zap.String("err", err.Error()))
	}

	stamps := stamp.New()

	return &App{
		GOpsKitApp: &common.GOpsKitApp{
			Name:       APP_NAME,
			Executor:   exec,
			KubeClient: kc,
			Logger:     logger,
			Platform:   platform,
			Stamps:     stamps,
		},
	}
}
