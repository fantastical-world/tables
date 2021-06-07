package tables

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

//Table represents a table with meta data and rows
type Table struct {
	Meta Meta  `json:"meta"`
	Rows []Row `json:"rows"`
}

//Meta stores metadata for a table
type Meta struct {
	Name           string   `json:"name"`
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

//Backingstore represents a general contract needed for persisting tables.
type Backingstore interface {
	LoadTable(records [][]string, table string, rollExpression string) error
	GetTable(table string) ([][]string, error)
	TableExpression(expression string) ([][]string, error)
	RandomRow(table string) ([]string, int, error)
	GetRow(roll int, table string) ([]string, error)
	GetHeader(table string) ([]string, error)
	ListTables() ([]string, error)
	Delete(name string) error
	GetMeta(name string) (Meta, error)
}

func Load(records [][]string, name string, rollExpression string) (Table, error) {
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
			if rangedRoll(row[0]) {
				rollRange = row[0]
				//we will set dieRoll to the range start for sorting purposes
				parts := strings.Split(row[0], "-")
				dieRoll, _ = strconv.Atoi(parts[0])
			} else {
				dieRoll, err = strconv.Atoi(row[0])
				if err != nil {
					return Table{}, fmt.Errorf("first column must be an integer since it represents a die roll")
				}
			}
		}

		hasRollExpression := false
		for _, column := range row {
			if rollableString(column) {
				hasRollExpression = true
				break
			}
		}

		tableRow := Row{DieRoll: dieRoll, RollRange: rollRange, HasRollExpression: hasRollExpression, Results: row}
		table.Rows = append(table.Rows, tableRow)
	}

	table.Meta = Meta{Name: name, Headers: headers, ColumnCount: len(headers), RollableTable: rollable, RollExpression: rollExpression}

	return table, nil
}

//helpers

func rollableString(value string) bool {
	re := regexp.MustCompile(`{{\s*(?P<num>[0-9]*)[d](?P<sides>[0-9]+)(?P<mod>\+|-)?(?P<mod_num>[0-9]+)?\s*}}`)
	return re.MatchString(value)
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
