package app

import (
	"fmt"
	"github.com/fmjstudios/gopskit/pkg/core"
	"github.com/fmjstudios/gopskit/pkg/fs"
	"github.com/fmjstudios/gopskit/pkg/kv"
	"github.com/fmjstudios/gopskit/pkg/log"
	"github.com/fmjstudios/gopskit/pkg/proc"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"path/filepath"
	"time"

	"github.com/fmjstudios/gopskit/pkg/kube"
	"github.com/fmjstudios/gopskit/pkg/stamp"
	"github.com/hashicorp/vault-client-go"
)

const (
	Name         = "waltr"
	DefaultLabel = "app.kubernetes.io/name=vault"
)

// Opt is configuration option for the application State
type Opt func(a *State)

type CLIOpt func() func(a *State) *cobra.Command

// State is the implementation for the `waltr` command-line application state
type State struct {
	*core.API

	// VaultClient is the HashCorp first-party Go Vault HTTP client, which waltr
	// uses for nearly all of its functionality
	VaultClient *vault.Client
}

// New creates a newly initialized instance of the State type
func New(opts ...Opt) (*State, error) {
	var err error

	platf, err := fs.Paths(fs.WithAppName(Name))
	if err != nil {
		return nil, err
	}

	lgr := log.New(log.WithCustomConfig(zap.NewDevelopmentConfig()))
	defer func() {
		err = lgr.Sync()
	}()
	if err != nil {
		return nil, err
	}

	exec, err := proc.NewExecutor(proc.WithInheritedEnv())
	if err != nil {
		return nil, err
	}

	kc, err := kube.NewClient()
	if err != nil {
		return nil, fmt.Errorf("could not create kubernetes client: %v", err)
	}

	// embedded BadgerDB database
	dbpath := filepath.Join(platf.Data, "data")
	fmt.Println("Creating BadgerDB database at:", dbpath)
	db, err := kv.New(dbpath)
	if err != nil {
		return nil, err
	}

	// enforce HTTPS
	va := fmt.Sprintf("http://127.0.0.1:%s", kube.DefaultLocalPort)
	vc, err := vault.New(vault.WithAddress(va), vault.WithRequestTimeout(60*time.Second), vault.WithTLS(vault.TLSConfiguration{
		InsecureSkipVerify: true,
	}))

	if err != nil {
		return nil, fmt.Errorf("could not create vault client: %v", err)
	}
	stamps := stamp.New()

	a := &State{
		API: &core.API{
			Name:  Name,
			Exec:  exec,
			Kube:  kc,
			Log:   lgr,
			KV:    db,
			Paths: platf,
			Stamp: stamps,
		},
		VaultClient: vc,
	}

	// (re-)configure if the user wants to do so
	for _, o := range opts {
		o(a)
	}

	return a, nil
}

// WithVaultOpts configures waltr's VaultClient instance with custom Options
// from the vault-client-go package
func WithVaultOpts(opts ...vault.ClientOption) Opt {
	return func(a *State) {
		a.VaultClient = proc.Must(vault.New(opts...))
	}
}
