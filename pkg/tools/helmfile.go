package tools

import (
	"github.com/helmfile/helmfile/pkg/app"
	"github.com/helmfile/helmfile/pkg/config"
)

// WriteValues implements Helmfile's 'write-values' command
func WriteValues(global *config.GlobalImpl) error {
	writeValuesOptions := config.NewWriteValuesOptions()
	impl := config.NewWriteValuesImpl(global, writeValuesOptions)

	h := app.New(impl)
	h.WriteValues(impl)
	return nil
}
