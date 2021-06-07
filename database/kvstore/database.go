package kvstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fantastical-world/dice"
	"github.com/fantastical-world/tables"

	"github.com/boltdb/bolt"
)

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

//LoadTable will load a table with the CSV data. All existing data in the table will be replaced.
func (d *Database) LoadTable(records [][]string, table string, rollExpression string) error {
	if rollExpression == "" {
		return d.loadStandardTable(records, table)
	}

	return d.loadRollableTable(records, table, rollExpression)
}

func (d *Database) loadStandardTable(records [][]string, table string) error {
	d.Lock()
	defer d.Unlock()
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b != nil {
			//should not need to worry about error since we are in here if a bucket exists
			_ = tx.DeleteBucket([]byte(table))
		}

		b, err = tx.CreateBucket([]byte(table))
		if err != nil {
			return err
		}

		var headers []string
		for i, line := range records {

			if i == 0 {
				headers = append(headers, line...)
				continue
			}

			hasRollExpression := false
			for _, value := range line {
				if RollableString(value) {
					hasRollExpression = true
					break
				}
			}

			//for standard tables we don't have a "dieRoll", but we do use the row number here for sorting purposes.
			row := tables.Row{DieRoll: i, RollRange: "", HasRollExpression: hasRollExpression, Results: line}
			encodedRow, _ := json.Marshal(row)

			err = b.Put([]byte(line[0]), encodedRow)
			if err != nil {
				return err
			}
		}

		meta := tables.Meta{Name: table, Headers: headers, ColumnCount: len(headers), RollableTable: false, RollExpression: ""}
		encoded, err := json.Marshal(meta)
		if err != nil {
			return err
		}

		err = b.Put([]byte("tables.meta"), encoded)
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

func (d *Database) loadRollableTable(records [][]string, table string, rollExpression string) error {
	d.Lock()
	defer d.Unlock()
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b != nil {
			//should not need to worry about error since we are in here if a bucket exists
			_ = tx.DeleteBucket([]byte(table))
		}

		b, err = tx.CreateBucket([]byte(table))
		if err != nil {
			return err
		}

		var headers []string
		for i, line := range records {

			if i == 0 {
				headers = append(headers, line...)
				continue
			}

			dieRoll := 0
			rollRange := ""

			if rangedRoll(line[0]) {
				rollRange = line[0]
				//we will set dieRoll to the range start for sorting purposes
				parts := strings.Split(line[0], "-")
				dieRoll, _ = strconv.Atoi(parts[0])
			} else {
				dieRoll, err = strconv.Atoi(line[0])
				if err != nil {
					return fmt.Errorf("first column must be an integer since it represents a die roll")
				}
			}

			hasRollExpression := false
			//check looks odd but once we find a row with at least one rollable string we won't bother checking the remainder
			for _, value := range line {
				if RollableString(value) {
					hasRollExpression = true
					break
				}
			}

			row := tables.Row{DieRoll: dieRoll, RollRange: rollRange, HasRollExpression: hasRollExpression, Results: line}
			encodedRow, err := json.Marshal(row)
			if err != nil {
				return err
			}

			err = b.Put([]byte(line[0]), encodedRow)
			if err != nil {
				return err
			}
		}

		meta := tables.Meta{Name: table, Headers: headers, ColumnCount: len(headers), RollableTable: true, RollExpression: rollExpression}
		encoded, err := json.Marshal(meta)
		if err != nil {
			return err
		}

		err = b.Put([]byte("tables.meta"), encoded)
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

//GetTable will return all the table rows
func (d *Database) GetTable(table string) ([][]string, error) {
	d.Lock()
	defer d.Unlock()
	var data [][]string
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return nil, err
	}
	defer db.Close()
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b == nil {
			return fmt.Errorf("table [%s] does not exist", table)
		}
		c := b.Cursor()
		meta := b.Get([]byte("tables.meta"))
		if meta != nil {
			var decoded tables.Meta
			err := json.Unmarshal(meta, &decoded)
			if err != nil {
				return err
			}
			data = append(data, decoded.Headers)
		}
		var rows []tables.Row
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if string(k) == "tables.meta" {
				continue
			}
			var decoded tables.Row
			err := json.Unmarshal(v, &decoded)
			if err != nil {
				return err
			}
			rows = append(rows, decoded)
		}

		sort.SliceStable(rows, func(i, j int) bool {
			return rows[i].DieRoll < rows[j].DieRoll
		})

		for _, row := range rows {
			data = append(data, row.Results)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return data, nil
}

//TableExpression returns table rows based on expression
func (d *Database) TableExpression(expression string) ([][]string, error) {
	wantsUnique := false
	if strings.HasPrefix(expression, "uni:") {
		wantsUnique = true
		expression = strings.ReplaceAll(expression, "uni:", "")
	}

	var data [][]string
	//simple 1d4+1 style expressions for tables (n?table or n#table)
	re := regexp.MustCompile(`^(?P<num>[0-9]*)([\?|#])(?P<table>[a-zA-Z,0-9,_,\.,\-]+)$`)
	if !re.MatchString(expression) {
		return nil, fmt.Errorf("not a valid table expression, must be ?table or n?table or n#table (e.g. ?npc, 2?npc, 3#npc)")
	}

	match := re.FindStringSubmatch(expression)
	request := match[2]
	number, _ := strconv.Atoi(match[1])
	if number == 0 && request == "?" {
		number = 1
	}
	if number == 0 && request == "#" {
		return nil, fmt.Errorf("not a valid table expression, a request to show a specific row must include a row number")
	}

	tableName := match[3]
	meta, err := d.GetMeta(tableName)
	if err != nil {
		return nil, err
	}

	if !meta.RollableTable {
		return nil, fmt.Errorf("not a rollable table, no roll expression available")
	}

	header, err := d.GetHeader(tableName)
	if err != nil {
		return nil, err
	}
	data = append(data, header)
	if request == "?" {
		if wantsUnique {
			var previousRolls []int
			for i := 0; i < number; i++ {
				row, roll, err := d.RandomRow(tableName)
				if containsRoll(previousRolls, roll) {
					i--
					continue
				}
				previousRolls = append(previousRolls, roll)
				if err != nil {
					return nil, err
				}
				data = append(data, row)
			}
			return data, nil
		}

		for i := 0; i < number; i++ {
			row, _, err := d.RandomRow(tableName)
			if err != nil {
				return nil, err
			}
			data = append(data, row)
		}
		return data, nil
	}

	row, err := d.GetRow(number, tableName)
	if err != nil {
		return nil, err
	}

	data = append(data, row)
	return data, nil
}

//RandomRow returns a random row entry from the specified table and it's roll value.
func (d *Database) RandomRow(table string) ([]string, int, error) {
	d.Lock()
	rollExpression := ""
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		d.Unlock()
		return nil, 0, err
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b == nil {
			return fmt.Errorf("table [%s] does not exist", table)
		}

		meta := b.Get([]byte("tables.meta"))
		if meta == nil {
			return errors.New("metadata does not exist for table")
		}

		var decoded tables.Meta
		err := json.Unmarshal(meta, &decoded)
		if err != nil {
			return err
		}

		rollExpression = decoded.RollExpression
		return nil
	})
	if err != nil {
		errC := db.Close()
		if errC != nil {
			d.Unlock()
			return nil, 0, errC
		}
		d.Unlock()
		return nil, 0, err
	}
	err = db.Close()
	if err != nil {
		d.Unlock()
		return nil, 0, err
	}
	d.Unlock()

	_, dieRoll, _ := dice.RollExpression(rollExpression)
	row, err := d.GetRow(dieRoll, table)
	if err != nil {
		return nil, 0, err
	}
	return row, dieRoll, nil
}

//GetRow returns the row entry from the specified table based on the roll value.
func (d *Database) GetRow(roll int, table string) ([]string, error) {
	d.Lock()
	defer d.Unlock()
	var row []string
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return nil, err
	}
	defer db.Close()
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b == nil {
			return fmt.Errorf("table [%s] does not exist", table)
		}

		key := strconv.Itoa(roll)
		value := b.Get([]byte(key))
		if value == nil {
			//let's check for any ranged rows that may match
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				if string(k) == "tables.meta" || !rangedRoll(string(k)) {
					continue
				}

				if rollInRange(roll, string(k)) {
					value = v
					break
				}
			}

			if value == nil {
				return fmt.Errorf("value for [%d] does not exist", roll)
			}
		}

		var decoded tables.Row
		err := json.Unmarshal(value, &decoded)
		if err != nil {
			return err
		}

		if decoded.HasRollExpression {
			var rolledResults []string
			for _, result := range decoded.Results {
				rolledResults = append(rolledResults, rollString(result))
			}
			row = rolledResults
		} else {
			row = decoded.Results
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return row, nil
}

//GetHeader returns the header of a table
func (d *Database) GetHeader(table string) ([]string, error) {
	d.Lock()
	defer d.Unlock()
	var header []string
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return nil, err
	}
	defer db.Close()
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b == nil {
			return fmt.Errorf("table [%s] does not exist", table)
		}
		meta := b.Get([]byte("tables.meta"))
		if meta != nil {
			var decoded tables.Meta
			err := json.Unmarshal(meta, &decoded)
			if err != nil {
				return err
			}
			header = decoded.Headers
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return header, nil
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
	err = db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			meta := b.Get([]byte("tables.meta"))
			if meta != nil {
				var decoded tables.Meta
				err := json.Unmarshal(meta, &decoded)
				if err != nil {
					return err
				}
				tableData = append(tableData, fmt.Sprintf("%s,%s,%t", decoded.Name, decoded.RollExpression, decoded.RollableTable))
			}
			return nil
		})
	})

	if err != nil {
		return nil, err
	}
	return tableData, nil
}

//Delete will delete a table.
func (d *Database) Delete(name string) error {
	d.Lock()
	defer d.Unlock()
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		err = tx.DeleteBucket([]byte(name))
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

//GetMeta returns a table's meta data.
func (d *Database) GetMeta(name string) (tables.Meta, error) {
	d.Lock()
	defer d.Unlock()
	var meta tables.Meta
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return meta, err
	}
	defer db.Close()
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		if b == nil {
			return fmt.Errorf("table [%s] does not exist", name)
		}
		m := b.Get([]byte("tables.meta"))
		if m != nil {
			err := json.Unmarshal(m, &meta)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return meta, err
	}

	return meta, nil
}

//RollableString validates a string to see if it is a valid roll expression
func RollableString(value string) bool {
	re := regexp.MustCompile(`{{\s*(?P<num>[0-9]*)[d](?P<sides>[0-9]+)(?P<mod>\+|-)?(?P<mod_num>[0-9]+)?\s*}}`)
	return re.MatchString(value)
}

func rollString(value string) string {
	rolledValue := value
	re := regexp.MustCompile(`{{\s*(?P<num>[0-9]*)[d](?P<sides>[0-9]+)(?P<mod>\+|-)?(?P<mod_num>[0-9]+)?\s*}}`)
	if !re.MatchString(value) {
		return value
	}

	match := re.FindAllStringSubmatch(value, 99) //limit to 99 rolls per value
	for _, m := range match {
		expression := strings.ReplaceAll(m[0], "{{", "")
		expression = strings.ReplaceAll(expression, "}}", "")
		_, sum, _ := dice.RollExpression(strings.Trim(expression, " "))
		rolledValue = strings.Replace(rolledValue, m[0], strconv.Itoa(sum), 1)
	}

	return rolledValue
}

func rangedRoll(value string) bool {
	if !strings.Contains(value, "-") {
		return false
	}

	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return false
	}

	_, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}

	_, err = strconv.Atoi(parts[1])
	return err == nil
}

func rollInRange(value int, rollRange string) bool {
	//shouldn't need this check since I'll only call with a valid range
	if !rangedRoll(rollRange) {
		return false
	}

	parts := strings.Split(rollRange, "-")
	start, _ := strconv.Atoi(parts[0])
	end, _ := strconv.Atoi(parts[1])

	if value >= start && value <= end {
		return true
	}

	return false
}

func containsRoll(i []int, roll int) bool {
	for _, v := range i {
		if v == roll {
			return true
		}
	}

	return false
}
