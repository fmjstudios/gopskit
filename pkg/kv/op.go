package kv

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/fmjstudios/gopskit/pkg/helpers"
	"golang.org/x/sync/errgroup"
	"strings"
	"time"
)

// enforce implementation of the interface
var _ Store = (*Database)(nil)

// Get implements the Store interface for Database. It retrieves a key from the database with
// the given OperationOpt options to configure the current operation. The operation itself is
// handled asynchronously, although the method itself is also thread-safe.
func (d *Database) Get(key string) (value []byte, err error) {
	// ensure we're getting clean keys
	err = d.ensureNonNamespaced(key)
	if err != nil {
		return nil, err
	}
	k := []byte(key)

	var bytes []byte
	g := new(errgroup.Group)
	g.Go(func() error {
		value, err := d.get(k)
		if err != nil {
			return err
		}

		bytes = append(bytes, value...)
		return nil
	})

	err = g.Wait()
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// get is the actual implementation of the retrieval of a value from the BadgerDB
// database. The key itself is namespaced beforehand to allow for multiple data
// models to be persisted simultaneously.
func (d *Database) get(key []byte) (value []byte, err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	var bytes []byte
	k := d.namespace(d.currentNamespace, string(key))
	err = d.kv.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err != nil {
			return err
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		bytes = append(bytes, value...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// Set implements the Store interface for Database. It sets a key within the database to
// a certain value with the given OperationOpt options to configure the current operation.
// The operation itself is handled asynchronously, although the method itself is also thread-safe.
func (d *Database) Set(key string, value []byte) error {
	err := d.ensureNonNamespaced(key)
	if err != nil {
		return err
	}
	k := []byte(key)

	g := new(errgroup.Group)
	g.Go(func() error {
		err := d.set(k, value)
		if err != nil {
			return err
		}

		return nil
	})

	err = g.Wait()
	if err != nil {
		return err
	}

	return nil
}

// get is the actual implementation of setting a value within the BadgerDB
// database. The key itself is namespaced beforehand to allow for multiple data
// models to be persisted simultaneously.
func (d *Database) set(key, value []byte) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	k := d.namespace(d.currentNamespace, string(key))
	var err error
	err = d.kv.Update(func(txn *badger.Txn) error {
		err := txn.Set(k, value)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// Has checks if the database contains a key by trying to read the value at the
// given key via the use of Get. If the method returns an error it is determined
// that the value does not exist and the error from Get is propagated alongside
// a false return value. Otherwise,  it does and a nil-error is returned.
func (d *Database) Has(key string) (bool, error) {
	err := d.ensureNonNamespaced(key)
	if err != nil {
		return false, err
	}
	k := []byte(key)
	value, err := d.get(k)
	if err != nil {
		return false, err
	}

	return len(value) > 0, nil
}

// Namespaces returns the current namespaces the database has been initialized with.
// The function cannot error since by default only the (single) "default" namespace
// will be returned if the DB is otherwise largely unconfigured.
func (d *Database) Namespaces() []string {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.namespaces
}

// Delete deletes a key from the database
func (d *Database) Delete(key string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	err := d.delete([]byte(key))
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) delete(key []byte) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	k := d.namespace(d.currentNamespace, string(key))
	var err error
	err = d.kv.Update(func(txn *badger.Txn) error {
		err := txn.Delete(k)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// Close closes an active connection to a database. This method must be called
// during program shutdown or else you risk corrupting the data. Additionally,
// we run garbage collection across the entire database before finally shutting
// down.
func (d *Database) Close() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	doneC := make(chan struct{}, 1)
	go d.gc(doneC)
	<-doneC

	return d.kv.Close()
}

// Path returns the filesystem path to the database
func (d *Database) Path() string {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.path
}

// Config returns the badger.Options the database was initialized with
func (d *Database) Config() badger.Options {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.options
}

// namespace namespaces a given key by prefixing the value with a namespace like "namespace/value".
// If an empty string is passed as the namespace, the DefaultNamespace "default" is used instead.
func (d *Database) namespace(namespace, key string) []byte {
	if namespace == "" {
		namespace = DefaultNamespace
	}

	prefix := fmt.Sprintf("%s/", namespace)
	return []byte(prefix + key)
}

// ensure non-namespaced ensures that a given key value contains no slashes and consists only of
// letters, thereby equating to a valid key.
func (d *Database) ensureNonNamespaced(key string) error {
	if strings.Contains(key, "/") && helpers.OnlyLetters(key) {
		return fmt.Errorf("cannot set value for a namespaced key. please exclude namespaces from the key")
	}

	return nil
}

// gc runs the garbage-collection for the BadgerDB which saves filesystem space. It is run
// as a goroutine before closing the database connection.
func (d *Database) gc(doneChan chan struct{}) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
	again:
		err := d.kv.RunValueLogGC(0.7)
		if err == nil {
			goto again
		}
	}

	doneChan <- struct{}{}
}
