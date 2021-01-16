package kvstore

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"fantastical.world/dice"
	"fantastical.world/tables"
	"github.com/boltdb/bolt"
)

//Database is a simple representation of a table database.
type Database struct {
	dbLocation string
	timeout    time.Duration
}

//this does nothing more than validate Backingstore interface compliance
var _ tables.Backingstore = (*Database)(nil)

//New creates a new Database for persisting and working with tables.
func New(dbLocation string) (tables.Backingstore, error) {
	db := Database{dbLocation: dbLocation, timeout: time.Second * 10}
	err := db.Prepare()
	if err != nil {
		return db, err
	}

	return db, nil
}

//Prepare noop
func (d Database) Prepare() error {
	return nil
}

//LoadTable will load a table with the CSV data. All existing data in the table will be replaced.
func (d Database) LoadTable(csvFile string, table string, rollExpression string) error {
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		records, err := readCSV(csvFile)
		if err != nil {
			return err
		}

		b := tx.Bucket([]byte(table))
		if b != nil {
			err = tx.DeleteBucket([]byte(table))
			if err != nil {
				return err
			}
		}

		b, err = tx.CreateBucket([]byte(table))
		if err != nil {
			return err
		}

		var headers []string
		tableType := tables.Simple

		for i, line := range records {

			if i == 0 {
				for _, header := range line {
					headers = append(headers, header)
				}
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

			//check looks odd but once we find a row with at least one rollable string we won't bother checking the remainder
			if tableType != tables.Advanced {
				for _, value := range line {
					if RollableString(value) {
						tableType = tables.Advanced
						break
					}
				}
			}

			row := tables.Row{DieRoll: dieRoll, RollRange: rollRange, Results: line}
			encodedRow, err := json.Marshal(row)
			if err != nil {
				return err
			}

			err = b.Put([]byte(line[0]), encodedRow)
			if err != nil {
				return err
			}
		}

		meta := tables.Meta{Type: tableType, Name: table, Headers: headers, ColumnCount: len(headers), RollExpression: rollExpression}
		encoded, err := json.Marshal(meta)
		if err != nil {
			return err
		}

		err = b.Put([]byte("meta"), encoded)
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

//AppendToTable will append CSV data to a table.
func (d Database) AppendToTable(csvFile string, table string, rollExpression string) error {
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		records, err := readCSV(csvFile)
		if err != nil {
			return err
		}

		b := tx.Bucket([]byte(table))
		if b == nil {
			return fmt.Errorf("can not append to table [%s] since does not exist, try running again without -a", table)
		}

		meta := b.Get([]byte("meta"))
		if meta == nil {
			return errors.New("metadata does not exist for table")
		}

		var decoded tables.Meta
		err = json.Unmarshal(meta, &decoded)
		if err != nil {
			return err
		}

		for i, line := range records {
			if i == 0 {
				if len(line) != decoded.ColumnCount {
					return fmt.Errorf("did not append data because column counts do not match, check your CSV")
				}
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

			//check looks odd but once we find a row with at least one rollable string we won't bother checking the remainder
			if decoded.Type != tables.Advanced {
				for _, value := range line {
					if RollableString(value) {
						decoded.Type = tables.Advanced
						break
					}
				}
			}

			row := tables.Row{DieRoll: dieRoll, RollRange: rollRange, Results: line}
			encodedRow, err := json.Marshal(row)
			if err != nil {
				return err
			}

			err = b.Put([]byte(line[0]), encodedRow)
			if err != nil {
				return err
			}
		}

		decoded.RollExpression = rollExpression
		encoded, err := json.Marshal(decoded)
		if err != nil {
			return err
		}

		err = b.Put([]byte("meta"), encoded)
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
func (d Database) GetTable(table string) ([][]string, error) {
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
		meta := b.Get([]byte("meta"))
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
			if string(k) == "meta" {
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
func (d Database) TableExpression(expression string) ([][]string, error) {
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
func (d Database) RandomRow(table string) ([]string, int, error) {
	rollExpression := ""
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return nil, 0, err
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b == nil {
			return fmt.Errorf("table [%s] does not exist", table)
		}

		meta := b.Get([]byte("meta"))
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
		err = db.Close()
		if err != nil {
			return nil, 0, err
		}
		return nil, 0, err
	}
	err = db.Close()
	if err != nil {
		return nil, 0, err
	}

	_, dieRoll, _ := dice.RollExpression(rollExpression)
	row, err := d.GetRow(dieRoll, table)
	if err != nil {
		return nil, 0, err
	}
	return row, dieRoll, nil
}

//GetRow returns the row entry from the specified table based on the roll value.
func (d Database) GetRow(roll int, table string) ([]string, error) {
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
				if string(k) == "meta" || !rangedRoll(string(k)) {
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

		advanced := false
		meta := b.Get([]byte("meta"))
		if meta != nil {
			var decoded tables.Meta
			err := json.Unmarshal(meta, &decoded)
			if err != nil {
				return err
			}

			if decoded.Type == tables.Advanced {
				advanced = true
			}
		}

		var decoded tables.Row
		err := json.Unmarshal(value, &decoded)
		if err != nil {
			return err
		}

		if advanced {
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
func (d Database) GetHeader(table string) ([]string, error) {
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
		meta := b.Get([]byte("meta"))
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
func (d Database) ListTables() ([]string, error) {
	var tableData []string
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return nil, err
	}
	defer db.Close()
	err = db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			meta := b.Get([]byte("meta"))
			if meta != nil {
				var decoded tables.Meta
				err := json.Unmarshal(meta, &decoded)
				if err != nil {
					return err
				}
				tableData = append(tableData, fmt.Sprintf("%s,%s,%s", decoded.Name, decoded.RollExpression, decoded.Type))
			}
			return nil
		})
	})

	if err != nil {
		return nil, err
	}
	return tableData, nil
}

//WriteTable will write the table to a csv file
func (d Database) WriteTable(table string, filename string) error {
	data, err := d.GetTable(table)
	if err != nil {
		return err
	}
	err = writeCSV(filename, data)
	if err != nil {
		return err
	}

	return nil
}

//Delete will delete a table.
func (d Database) Delete(name string) error {
	db, err := bolt.Open(d.dbLocation, 0600, &bolt.Options{Timeout: d.timeout})
	if err != nil {
		return err
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte(name))
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

//RollableString validates a string to see if it is a valid roll expression
func RollableString(value string) bool {
	re := regexp.MustCompile(`{{\s*(?P<num>[0-9]*)[d](?P<sides>[0-9]+)(?P<mod>\+|-)?(?P<mod_num>[0-9]+)?\s*}}`)
	return re.MatchString(value)
}

func readCSV(filename string) ([][]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func writeCSV(filename string, data [][]string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)

	err = w.WriteAll(data)
	if err != nil {
		return err
	}

	return nil
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
		_, sum, err := dice.RollExpression(strings.Trim(expression, " "))
		if err != nil {
			return value
		}
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
	if err != nil {
		return false
	}

	return true
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
