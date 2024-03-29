package tables

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/fantastical-world/dice"
)

type TableError string

func (be TableError) Error() string { return string(be) }

const ErrTableDoesNotExist = TableError("table does not exist")
const ErrTableInvalid = TableError("table invalid")
const ErrInvalidTableRollValue = TableError("roll value is not valid for this table")
const ErrTableNotRollable = TableError("not a rollable table, no roll expression available")
const ErrInvalidTableExpression = TableError("not a valid table expression")
const ErrTableDoesNotMatchTableExpression = TableError("table is not the table in the table expression")
const ErrInvalidRollColumn = TableError("first column must be an integer since it represents a die roll")

var (
	TableRollExpressionRE = regexp.MustCompile(`^([0-9]*)([\?|#])([a-zA-Z,0-9,_,\.,\-]+)$`)
)

//Table represents a table with meta data and rows
type Table struct {
	Meta Meta  `json:"meta"`
	Rows []Row `json:"rows"`
}

//Meta stores metadata for a table
type Meta struct {
	Name           string   `json:"name"`
	DisplayName    string   `json:"display_name"`
	Title          string   `json:"title"`
	FlavorText     string   `json:"flavor_text"`
	Campaign       string   `json:"campaign"`
	Headers        []string `json:"headers"`
	ColumnCount    int      `json:"column_count"`
	RollableTable  bool     `json:"rollable_table"`
	RollExpression string   `json:"roll_expression"`
}

//Row represents a row from a table
type Row struct {
	DieRoll           int      `json:"die_roll"`
	RollRange         string   `json:"roll_range"`
	HasRollExpression bool     `json:"has_roll_expression"`
	Results           []string `json:"results"`
}

func (t Table) Pack() (string, []byte) {
	b, _ := json.Marshal(t)

	return t.Meta.Name, b
}

func (t *Table) Unpack(data []byte) {
	err := json.Unmarshal(data, &t)
	if err != nil {
		t = &Table{}
	}
}

func (t Table) Header() []string {
	return t.Meta.Headers
}

func (t Table) Records() [][]string {
	var records [][]string
	records = append(records, t.Meta.Headers)
	for _, row := range t.Rows {
		records = append(records, row.Results)
	}

	return records
}

func (t Table) RandomRow() ([]string, int, error) {
	dieRoll := 0

	if t.Meta.RollableTable {
		_, dieRoll, _ = dice.RollExpression(t.Meta.RollExpression)
	} else {
		//in the past we didn't allow random rows if table not rollable, but now we want to
		rollExpression := fmt.Sprintf("1d%d", len(t.Rows))
		_, dieRoll, _ = dice.RollExpression(rollExpression)
	}

	row, err := t.GetRow(dieRoll)
	if err != nil {
		return nil, 0, err
	}

	return row, dieRoll, nil
}

func (t Table) GetRow(roll int) ([]string, error) {
	for _, row := range t.Rows {
		if row.DieRoll == roll {
			if row.HasRollExpression {
				var rolledResults []string
				for _, result := range row.Results {
					rolledResults = append(rolledResults, dice.RollString(result))
				}
				return rolledResults, nil
			}
			return row.Results, nil
		}
	}

	//this means we didn't find a row with the roll requested, so let's check again with ranges
	for _, row := range t.Rows {
		if RollInRange(roll, row.RollRange) {
			if row.HasRollExpression {
				var rolledResults []string
				for _, result := range row.Results {
					rolledResults = append(rolledResults, dice.RollString(result))
				}
				return rolledResults, nil
			}
			return row.Results, nil
		}
	}

	return nil, ErrInvalidTableRollValue
}

func (t Table) Expression(te string) ([][]string, error) {
	if !t.Meta.RollableTable {
		return nil, ErrTableNotRollable
	}

	wantsUnique := false
	if strings.HasPrefix(te, "uni:") {
		wantsUnique = true
		te = strings.ReplaceAll(te, "uni:", "")
	}

	var data [][]string
	if !TableRollExpressionRE.MatchString(te) {
		return nil, ErrInvalidTableExpression
	}

	match := TableRollExpressionRE.FindStringSubmatch(te)
	request := match[2]
	number, _ := strconv.Atoi(match[1])
	if number == 0 && request == "?" {
		number = 1
	}
	if number == 0 && request == "#" {
		return nil, ErrInvalidTableExpression
	}

	tableName := match[3]
	if tableName != t.Meta.Name {
		return nil, ErrTableDoesNotMatchTableExpression
	}

	data = append(data, t.Meta.Headers)
	if request == "?" {
		if wantsUnique {
			var previousRolls []int
			for i := 0; i < number; i++ {
				row, roll, err := t.RandomRow()
				if err != nil {
					return nil, err
				}

				if containsRoll(previousRolls, roll) {
					//if the number of previousRolls matches length of available rolls we can no longer find unique rows
					if len(previousRolls) == len(t.Rows) {
						break
					}
					//let's keep trying
					i--
					continue
				}
				previousRolls = append(previousRolls, roll)
				data = append(data, row)
			}
			return data, nil
		}

		for i := 0; i < number; i++ {
			row, _, err := t.RandomRow()
			if err != nil {
				return nil, err
			}
			data = append(data, row)
		}

		return data, nil
	}

	row, err := t.GetRow(number)
	if err != nil {
		return nil, err
	}

	data = append(data, row)

	return data, nil
}

func (t Table) Hash() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(t.Meta.Name)))
}

//Load returns a Table loaded with the provided records as its rows. The first record will be used as its header.
//Providing a roll expression allow this table to be "rolled" using table expressions (e.g. 2?tablename, 4#tablename).
func Load(records [][]string, name, displayName, rollExpression string) (Table, error) {
	var headers []string
	var err error
	table := Table{}
	rollable := (rollExpression != "")

	for i, row := range records {
		if i == 0 {
			headers = append(headers, row...)
			continue
		}

		dieRoll := i
		rollRange := ""
		if rollable {
			dieRoll = 0
			if RangedRoll(row[0]) {
				rollRange = row[0]
				//we will set dieRoll to the range start for sorting purposes
				parts := strings.Split(row[0], "-")
				dieRoll, _ = strconv.Atoi(parts[0])
			} else {
				dieRoll, err = strconv.Atoi(row[0])
				if err != nil {
					return Table{}, ErrInvalidRollColumn
				}
			}
		}

		hasRollExpression := false
		for _, column := range row {
			if RollableString(column) {
				hasRollExpression = true
				break
			}
		}

		tableRow := Row{DieRoll: dieRoll, RollRange: rollRange, HasRollExpression: hasRollExpression, Results: row}
		table.Rows = append(table.Rows, tableRow)
	}

	table.Meta = Meta{Name: name, DisplayName: displayName, Headers: headers, ColumnCount: len(headers), RollableTable: rollable, RollExpression: rollExpression}

	return table, nil
}

//RollableString returns true if value contains a roll expression.
func RollableString(value string) bool {
	return dice.ContainsRollExpressionBracedRE.MatchString(value)
}

//RangedRoll returns true if value is a valid ranged roll.
//To be a valid ranged roll, value must be in #-# format (e.g. 1-6, 6-8).
func RangedRoll(value string) bool {
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

//RollInRange checks if the roll value is in the range provided.
func RollInRange(value int, rollRange string) bool {
	if !RangedRoll(rollRange) {
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

//ParseTablename returns the tablename from a table expression.
func ParseTablename(te string) string {
	if strings.HasPrefix(te, "uni:") {
		te = strings.ReplaceAll(te, "uni:", "")
	}

	if !TableRollExpressionRE.MatchString(te) {
		return ""
	}

	match := TableRollExpressionRE.FindStringSubmatch(te)
	return match[3]
}

func containsRoll(i []int, roll int) bool {
	for _, v := range i {
		if v == roll {
			return true
		}
	}

	return false
}
