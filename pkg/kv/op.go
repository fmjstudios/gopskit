package kv

import (
	"github.com/dgraph-io/badger/v4"
	"golang.org/x/sync/errgroup"
)

// Get retrieves the value at a specific key from the database
func (d *Database) Get(key string) (value string, err error) {
	buf := make([]byte, 0)
	keyB := []byte(key)
	g := new(errgroup.Group)

	g.Go(func() error {
		value, err := d.get(keyB)
		if err != nil {
			return err
		}

		buf = append(buf, value...)
		return nil
	})

	err = g.Wait()
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func (d *Database) get(key []byte) (value []byte, err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	buf := make([]byte, 0)
	err = d.kv.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		buf = append(buf, value...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return buf, nil
}

// Set sets the value at a specific key within the database
func (d *Database) Set(key string, value []byte) error {
	keyB := []byte(key)
	g := new(errgroup.Group)

	g.Go(func() error {
		err := d.set(keyB, value)
		if err != nil {
			return err
		}

		return nil
	})

	err := g.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) set(key, value []byte) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	var err error
	err = d.kv.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
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
