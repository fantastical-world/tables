package kvstore

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/fantastical-world/tables"

	"github.com/boltdb/bolt"
)

type BackingstoreError string

func (be BackingstoreError) Error() string { return string(be) }

const ErrTableDoesNotExist = BackingstoreError("table does not exist")

const TablesBucket = "__TABLES__"

//Database is a simple representation of a table database.
type Database struct {
	sync.Mutex
	dbLocation string
	timeout    time.Duration
}

//this does nothing more than validate Backingstore interface compliance
var _ tables.Backingstore = (*Database)(nil)

//New creates a new Database for persisting and working with tables.
func New(dbLocation string) tables.Backingstore {
	return &Database{dbLocation: dbLocation, timeout: time.Second * 10}
}

//SaveTable saves the provided table.
func (d *Database) SaveTable(table tables.Table) error {
	d.Lock()
	defer d.Unlock()
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(TablesBucket))
		encoded, _ := json.Marshal(table)

		err = b.Put([]byte(table.Meta.Name), encoded)
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

func (d *Database) GetTable(name string) (tables.Table, error) {
	var table tables.Table
	d.Lock()
	defer d.Unlock()
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return tables.Table{}, err
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(TablesBucket))
		bytes := b.Get([]byte(name))
		if bytes == nil {
			return ErrTableDoesNotExist
		}

		err := json.Unmarshal(bytes, &table)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return tables.Table{}, err
	}

	//not sure if needed
	sort.SliceStable(table.Rows, func(p, q int) bool {
		return table.Rows[p].DieRoll < table.Rows[q].DieRoll
	})

	return table, nil
}

func (d *Database) DeleteTable(name string) error {
	d.Lock()
	defer d.Unlock()
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(TablesBucket))
		bytes := b.Get([]byte(name))
		if bytes == nil {
			return ErrTableDoesNotExist
		}

		_ = b.Delete([]byte(name))

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

//ListTables will list all the tables and their metadata.
func (d *Database) ListTables() ([]string, error) {
	d.Lock()
	defer d.Unlock()
	var tableData []string
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return nil, err
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(TablesBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var table tables.Table
			err := json.Unmarshal(v, &table)
			if err != nil {
				return err
			}

			tableData = append(tableData, fmt.Sprintf("%s,%s,%t", table.Meta.Name, table.Meta.RollExpression, table.Meta.RollableTable))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.SliceStable(tableData, func(p, q int) bool {
		return tableData[p] < tableData[q]
	})

	return tableData, nil
}
