package filestore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"sync"

	"github.com/fantastical-world/tables"
)

type FileStore struct {
	sync.Mutex
	location string
	tables   map[string]tables.Table
}

var _ tables.Backingstore = (*FileStore)(nil)

//New creates a new FileStore for persisting and working with tables.
func New(location string) tables.Backingstore {
	tableData := make(map[string]tables.Table)
	_, err := os.Stat(location)
	if err == nil {
		data, err := ioutil.ReadFile(location)
		if err == nil {
			err = json.Unmarshal(data, &tableData)
			if err != nil {
				tableData = make(map[string]tables.Table)
			}
		}
	}

	return &FileStore{location: location, tables: tableData}
}

func (f *FileStore) SaveTable(table tables.Table) error {
	f.Lock()
	defer f.Unlock()
	if table.Meta.Name == "" {
		return tables.ErrTableInvalid
	}

	f.tables[table.Meta.Name] = table

	j, _ := json.Marshal(f.tables)

	fs, err := os.Create(f.location)
	if err != nil {
		return err
	}

	_, err = fs.Write(j)
	if err != nil {
		return err
	}

	return fs.Close()
}

func (f *FileStore) GetTable(name string) (tables.Table, error) {
	f.Lock()
	defer f.Unlock()
	table, exists := f.tables[name]
	if !exists {
		return tables.Table{}, tables.ErrTableDoesNotExist
	}
	return table, nil
}

func (f *FileStore) DeleteTable(name string) error {
	f.Lock()
	defer f.Unlock()
	_, exists := f.tables[name]
	if !exists {
		return tables.ErrTableDoesNotExist
	}

	delete(f.tables, name)

	j, _ := json.Marshal(f.tables)

	fs, err := os.Create(f.location)
	if err != nil {
		return err
	}

	_, err = fs.Write(j)
	if err != nil {
		return err
	}
	return fs.Close()
}

func (f *FileStore) ListTables() ([]string, error) {
	var tableData []string

	for _, v := range f.tables {
		tableData = append(tableData, fmt.Sprintf("%s,%s,%t", v.Meta.Name, v.Meta.RollExpression, v.Meta.RollableTable))
	}

	sort.SliceStable(tableData, func(p, q int) bool {
		return tableData[p] < tableData[q]
	})

	return tableData, nil
}
