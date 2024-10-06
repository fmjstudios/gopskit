package kv

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"sync"
)

const (
	DefaultKVName    = "store.db"
	DefaultNamespace = "access"
)

var (
	DefaultPath       = fmt.Sprintf("/tmp/%s", DefaultKVName)
	DefaultNamespaces = []string{DefaultNamespace}
)

// Database is the central type for the persisted Key-Value database
type Database struct {
	// Implementation
	//
	// kv is the underlying BadgerDB key-value store
	kv *badger.DB

	// conf refers to the latest 'version' of the current badger.Options for kv
	conf badger.Options

	// path is the filesystem path the database is persisted to
	path string

	// ns represents the namespaces managed by the Database
	ns []string

	// opt is the currently used reference to an Operation object which configures
	// the respective database Operation like Get, Set, Has, etc...
	opt *Operation

	// Asynchronicity
	//
	// ctx represents the embedded database context to cancel operations
	ctx context.Context

	// cancel is the context.CancelFunc associated with the embedded context
	cancel context.CancelFunc

	// lock is a Mutex which ensures that only one goroutine may write or reads
	// from the database at a time to prevent corruptions
	lock sync.Mutex
}

// Operation is the configuration object for database operations
type Operation struct {
	// namespace represents the namespace to operate on within the current database
	namespace string

	// format is the output Format to use for the specific value
	format Format

	// lock is a Mutex which ensures that only one goroutine may modify the configuration
	// at a time
	lock sync.Mutex
}

type Store interface {
	Get(key string, opts ...OperationOpt) (value []byte, err error)
	Set(key string, value []byte, opts ...OperationOpt) error
	Has(key string, opts ...OperationOpt) (contains bool, err error)
	Namespaces(opts ...OperationOpt) (namespaces []string, err error)
	Delete(key string, opts ...OperationOpt) error
	Close() error
	Path() string
	Config() badger.Options
}

// Opt is a configuration option for the newly initialized Key-Value database
type Opt func(k *Database)

// OperationOpt is configuration option for database commands like Get, Set, etc.
type OperationOpt func(o *Operation)
