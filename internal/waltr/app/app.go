package app

import (
	"fmt"
	"time"

	"github.com/fmjstudios/gopskit/pkg/cmd"
	"github.com/fmjstudios/gopskit/pkg/common"
	"github.com/fmjstudios/gopskit/pkg/kube"
	"github.com/fmjstudios/gopskit/pkg/logger"
	"github.com/fmjstudios/gopskit/pkg/platform"
	"github.com/fmjstudios/gopskit/pkg/stamp"
	"github.com/hashicorp/vault-client-go"
	"go.uber.org/zap"
)

const (
	Name             = "waltr"
	DefaultNamespace = "vault"
	DefaultLabel     = "app.kubernetes.io/name=vault"
)

type Opt func(a *App)

// WithVaultOpts configures waltr's VaultClient instance with custom Options
// from the vault-client-go package
func WithVaultOpts(opts ...vault.ClientOption) Opt {
	return func(a *App) {
		a.VaultClient = cmd.Must(vault.New(opts...))
	}
}

// type App is the implementation for the `waltr` command-line application
type App struct {
	*common.GOpsKitApp

	// VaultClient is the HashiCorp first-party Go Vault HTTP client, which waltr
	// uses for nearly all of it's functionality post high-availability setup
	VaultClient *vault.Client
}

// New creates a newly initialized instance of the App type
func New(opts ...Opt) *App {
	var err error

	platform := platform.New(platform.WithApp(Name))
	logger := logger.New()
	defer func() {
		err = logger.Sync()
	}()

	if err != nil {
		logger.Fatal("could not Sync logger", zap.Error(err))
	}

	exec := cmd.NewExecutor(cmd.WithInheritedEnv())
	kc, err := kube.NewClient()
	if err != nil {
		logger.Fatal("could not create Kubernetes Client", zap.Error(err))
	}

	va := fmt.Sprintf("http://127.0.0.1:%s", kube.DefualtLocalPort)
	vc, err := vault.New(vault.WithAddress(va), vault.WithRequestTimeout(60*time.Second))
	if err != nil {
		logger.Fatal("could not create Vault Client", zap.Error(err))
	}

	stamps := stamp.New()

	a := &App{
		GOpsKitApp: &common.GOpsKitApp{
			Name:       Name,
			Executor:   exec,
			KubeClient: kc,
			Logger:     logger,
			Platform:   platform,
			Stamps:     stamps,
		},
		VaultClient: vc,
	}

	// (re-)configure if the user wants do so
	for _, o := range opts {
		o(a)
	}

	return a
}
