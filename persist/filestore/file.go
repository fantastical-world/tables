package filestore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/fantastical-world/tables"
)

type FileStore struct {
	sync.Mutex
	location string
	tables   map[string]tables.Table
}

var _ tables.Backingstore = (*FileStore)(nil)

//New creates a new FileStore using the provided location to read/write table files.
//If location is a valid directory all JSON files in the directory will be read and
//loaded into a Table. Any unsuccessful loads will be ignored.
func New(location string) (tables.Backingstore, error) {
	tableData := make(map[string]tables.Table)
	entries, err := os.ReadDir(location)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			var table tables.Table
			contents, err := ioutil.ReadFile(entry.Name())
			if err == nil {
				err = json.Unmarshal(contents, &table)
				if err == nil {
					tableData[table.Meta.Name] = table
				}
			}
		}
	}

	return &FileStore{location: location, tables: tableData}, nil
}

//SaveTable will add it to the FileStore and persist it to a JSON file.
func (f *FileStore) SaveTable(table tables.Table) error {
	f.Lock()
	defer f.Unlock()
	if table.Meta.Name == "" {
		return tables.ErrTableInvalid
	}

	f.tables[table.Meta.Name] = table

	j, _ := json.Marshal(table)

	fs, err := os.Create(table.Hash() + ".json")
	if err != nil {
		return err
	}

	_, err = fs.Write(j)
	if err != nil {
		return err
	}

	return fs.Close()
}

//GetTable returns the specified table.
func (f *FileStore) GetTable(name string) (tables.Table, error) {
	f.Lock()
	defer f.Unlock()
	table, exists := f.tables[name]
	if !exists {
		return tables.Table{}, tables.ErrTableDoesNotExist
	}
	return table, nil
}

//DeleteTable will delete the table and JSON file.
func (f *FileStore) DeleteTable(name string) error {
	f.Lock()
	defer f.Unlock()
	table, exists := f.tables[name]
	if !exists {
		return tables.ErrTableDoesNotExist
	}

	delete(f.tables, name)

	return os.Remove(table.Hash() + ".json")
}

//ListTable returns a listing of all tables.
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
