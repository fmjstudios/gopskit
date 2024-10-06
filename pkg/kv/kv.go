// Package kv implements a persistent Key-Value store based on the open-source BadgerDB. The implementation supports
// namespacing keys and their values as well as using a global (shared) namespace. The entire storage system works
// asynchronously using mutexes to restrict access to a single goroutine at a time.
package kv

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"sync"
)

// New instantiates a new key-value store and configures it with the given Opt
// configuration options
func New(path string, opts ...Opt) (*Database, error) {
	ctx, cancel := context.WithCancel(context.TODO())
	bOpts := badger.DefaultOptions(path)

	db := &Database{
		kv:     nil,
		conf:   bOpts,
		path:   path,
		ns:     DefaultNamespaces,
		opt:    DefaultOperation(),
		ctx:    ctx,
		cancel: cancel,
		lock:   sync.Mutex{},
	}

	// (re)-configure
	var wg sync.WaitGroup
	wg.Add(len(opts))
	for _, opt := range opts {
		go func() {
			opt(db)
			wg.Done()
		}()
	}
	wg.Wait()

	// db
	bdb, err := badger.Open(db.conf)
	if err != nil {
		return nil, fmt.Errorf("could not open badger database: %w", err)
	}
	db.kv = bdb

	return db, nil
}

// DefaultOperation set's the initial default Operation values
func DefaultOperation() *Operation {
	return &Operation{
		namespace: DefaultNamespaces[0],
		format:    YAML,
		lock:      sync.Mutex{},
	}
}

// SetNamespace ...
func (d *Database) SetNamespace(ns string) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.opt.namespace = ns
}

// SetFormat ...
func (d *Database) SetFormat(format Format) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.opt.format = format
}

//
// Instantiation Options
//

// WithPath ...
func WithPath(path string) Opt {
	return func(d *Database) {
		d.lock.Lock()
		defer d.lock.Unlock()

		d.path = path
	}
}

// WithExtraNamespaces ...
func WithExtraNamespaces(namespaces ...string) Opt {
	return func(d *Database) {
		d.lock.Lock()
		defer d.lock.Unlock()

		d.ns = append(d.ns, namespaces...)
	}
}

// WithBadgerOptions ...
func WithBadgerOptions(opts badger.Options) Opt {
	return func(d *Database) {
		d.lock.Lock()
		defer d.lock.Unlock()

		d.conf = opts
	}
}

// WithContext ...
func WithContext(ctx context.Context) Opt {
	return func(d *Database) {
		d.lock.Lock()
		defer d.lock.Unlock()

		nCTX, cancel := context.WithCancel(ctx)
		d.ctx = nCTX
		d.cancel = cancel
	}
}

//
// Operation Options
//

// WithNamespace ...
func WithNamespace(namespace string) OperationOpt {
	return func(o *Operation) {
		o.lock.Lock()
		defer o.lock.Unlock()

		o.namespace = namespace
	}
}

// WithFormat ...
func WithFormat(format Format) OperationOpt {
	return func(o *Operation) {
		o.lock.Lock()
		defer o.lock.Unlock()

		o.format = format
	}
}
