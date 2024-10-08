package helpers

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"fmt"
	"github.com/fmjstudios/gopskit/pkg/proc"
	"golang.org/x/sync/errgroup"
	"math/rand/v2"
	"os/exec"
)

type PassphraseConfig struct {
	Length  int
	CharSet string
}

type PassphraseOpt func(conf *PassphraseConfig)

func WithLength(length int) func(conf *PassphraseConfig) {
	return func(cfg *PassphraseConfig) {
		cfg.Length = length
	}
}

func WithCharSet(charset string) func(conf *PassphraseConfig) {
	return func(cfg *PassphraseConfig) {
		cfg.CharSet = charset
	}
}

func GeneratePassphrase(opts ...PassphraseOpt) string {
	var pass string

	cfg := &PassphraseConfig{
		Length:  48,
		CharSet: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789",
	}

	// configure
	for _, o := range opts {
		o(cfg)
	}

	for i := 1; i < cfg.Length; i++ {
		pass += string(cfg.CharSet[rand.IntN(len(cfg.CharSet))])
	}

	return pass
}

type Encoding int

const (
	Base64 Encoding = iota + 1
	Raw
)

type DiffieHellmanConfig struct {
	Bits     int
	Encoding Encoding
}

type DiffieHellmanOpt func(cfg *DiffieHellmanConfig)

func WithBits(bits int) func(cfg *DiffieHellmanConfig) {
	return func(cfg *DiffieHellmanConfig) {
		cfg.Bits = bits
	}
}

func WithEncoding(encoding Encoding) func(cfg *DiffieHellmanConfig) {
	return func(cfg *DiffieHellmanConfig) {
		cfg.Encoding = encoding
	}
}

func GenerateDiffieHellmanParams(opts ...DiffieHellmanOpt) (string, error) {
	var params string
	errg, _ := errgroup.WithContext(context.Background())

	// sanity
	_, err := exec.LookPath("openssl")
	if err != nil {
		return "", fmt.Errorf("openssl is not installed on the system")
	}

	cfg := &DiffieHellmanConfig{
		Bits:     4096,
		Encoding: Raw,
	}

	for _, o := range opts {
		o(cfg)
	}

	args := []string{"openssl", "dhparam", fmt.Sprintf("%d", cfg.Bits)}
	var bufStdO, bufStdE bytes.Buffer

	e, err := proc.NewExecutor(proc.WithInheritedEnv())
	if err != nil {
		return "", err
	}

	errg.Go(func() error {
		_, err := e.Execute(args, proc.WithWriters(bufStdO, bufStdE))
		if err != nil {
			return err
		}

		return nil
	})

	if err := errg.Wait(); err != nil {
		return "", proc.ExecuteError{
			ExitCode: e.ProcessState.ExitCode(),
			Err:      err,
		}
	}

	switch cfg.Encoding {
	case Base64:
		params = b64.StdEncoding.EncodeToString(bufStdO.Bytes())
	case Raw:
		params = bufStdO.String()
	default:
		return "", fmt.Errorf("invalid DiffieHellman encoding: %v", cfg.Encoding)
	}

	return params, nil
}
