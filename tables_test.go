package tables

import (
	"reflect"
	"testing"

	"github.com/fantastical-world/dice"
)

var testCSV = [][]string{
	{"D6", "Result", "Description"},
	{"1", "Fight {{1d1}} rats", "The party runs across some dirty rats."},
	{"2", "No encounter", "Nothing to see here."},
	{"3", "A wolf can be heard nearby", "If the party is careful they may avoid the wolf."},
	{"4", "{{1d1+1}} bats attack", "Angry bats swarm and attack the party."},
	{"5", "I can see you, can you see me?", "A whisper can be heard in the trees."},
	{"6", "A pile of bones covers {{1d1}}GP", "You found some loot."},
}

var rangedCSV = [][]string{
	{"D6", "Result"},
	{"1-2", "You rolled a 1 or 2"},
	{"3-4", "You rolled a 3 or 4"},
	{"5-6", "You rolled a 5 or 6"},
}

var rangedWithExpressionCSV = [][]string{
	{"D6", "Result"},
	{"1-2", "You rolled a 1 or 2"},
	{"3-4", "You rolled a 3 or 4, bonus {{1d1+1}}"},
	{"5-6", "You rolled a 5 or 6"},
}

var badCSV = [][]string{
	{"D3", "Result", "Description"},
	{"A", "Fight {{1d1}} rats", "The party runs across some dirty rats."},
	{"2:8", "No encounter", "Nothing to see here."},
	{"3", "A wolf can be heard nearby", "If the party is careful they may avoid the wolf."},
}

func Test_Load(t *testing.T) {
	var table Table
	var err error

	t.Run("validate loaded table, and meta data...", func(t *testing.T) {
		table, err = Load(testCSV, "test", "d6")
		if err != nil {
			t.Errorf("unexpected error, %s", err)
		}

		meta := Meta{Name: "test", Headers: []string{"D6", "Result", "Description"}, ColumnCount: 3, RollableTable: true, RollExpression: "d6"}
		if !reflect.DeepEqual(meta, table.Meta) {
			t.Errorf("want %v, got %v", meta, table.Meta)
		}
	})

	testCases := []struct {
		name  string
		index int
		want  Row
	}{
		{
			name:  "validate loaded table row 1...",
			index: 0,
			want:  Row{DieRoll: 1, RollRange: "", HasRollExpression: true, Results: []string{"1", "Fight {{1d1}} rats", "The party runs across some dirty rats."}},
		},
		{
			name:  "validate loaded table row 2...",
			index: 1,
			want:  Row{DieRoll: 2, RollRange: "", HasRollExpression: false, Results: []string{"2", "No encounter", "Nothing to see here."}},
		},
		{
			name:  "validate loaded table row 3...",
			index: 2,
			want:  Row{DieRoll: 3, RollRange: "", HasRollExpression: false, Results: []string{"3", "A wolf can be heard nearby", "If the party is careful they may avoid the wolf."}},
		},
		{
			name:  "validate loaded table row 4...",
			index: 3,
			want:  Row{DieRoll: 4, RollRange: "", HasRollExpression: true, Results: []string{"4", "{{1d1+1}} bats attack", "Angry bats swarm and attack the party."}},
		},
		{
			name:  "validate loaded table row 5...",
			index: 4,
			want:  Row{DieRoll: 5, RollRange: "", HasRollExpression: false, Results: []string{"5", "I can see you, can you see me?", "A whisper can be heard in the trees."}},
		},
		{
			name:  "validate loaded table row 6...",
			index: 5,
			want:  Row{DieRoll: 6, RollRange: "", HasRollExpression: true, Results: []string{"6", "A pile of bones covers {{1d1}}GP", "You found some loot."}},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			got := table.Rows[test.index]

			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("want %v, got %v", test.want, got)
			}
		})
	}
}

func Test_Load_Ranged(t *testing.T) {
	var table Table
	var err error

	t.Run("validate loaded table, and meta data...", func(t *testing.T) {
		table, err = Load(rangedCSV, "ranged", "d6")
		if err != nil {
			t.Errorf("unexpected error, %s", err)
		}

		meta := Meta{Name: "ranged", Headers: []string{"D6", "Result"}, ColumnCount: 2, RollableTable: true, RollExpression: "d6"}
		if !reflect.DeepEqual(meta, table.Meta) {
			t.Errorf("want %v, got %v", meta, table.Meta)
		}
	})

	testCases := []struct {
		name  string
		index int
		want  Row
	}{
		{
			name:  "validate row 1...",
			index: 0,
			want:  Row{DieRoll: 1, RollRange: "1-2", HasRollExpression: false, Results: []string{"1-2", "You rolled a 1 or 2"}},
		},
		{
			name:  "validate row 2...",
			index: 1,
			want:  Row{DieRoll: 3, RollRange: "3-4", HasRollExpression: false, Results: []string{"3-4", "You rolled a 3 or 4"}},
		},
		{
			name:  "validate row 3...",
			index: 2,
			want:  Row{DieRoll: 5, RollRange: "5-6", HasRollExpression: false, Results: []string{"5-6", "You rolled a 5 or 6"}},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			got := table.Rows[test.index]

			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("want %v, got %v", test.want, got)
			}
		})
	}
}

func Test_Load_Error(t *testing.T) {
	t.Run("validate an empty table and error is returned when data is invalid...", func(t *testing.T) {
		table, err := Load(badCSV, "bad", "d6")
		if err == nil {
			t.Error("expected error, got nil")
		}

		want := Table{}

		if !reflect.DeepEqual(want, table) {
			t.Errorf("want %v, got %v", want, table)
		}
	})
}

func TestTable_Header(t *testing.T) {
	t.Run("validate the header returned matched the loaded header...", func(t *testing.T) {
		table, err := Load(testCSV, "test", "d6")
		if err != nil {
			t.Errorf("unexpected error, %s", err)
		}

		got := table.Header()

		if !reflect.DeepEqual(testCSV[0], got) {
			t.Errorf("want %v, got %v", testCSV[0], got)
		}
	})
}

func TestTable_Records(t *testing.T) {
	t.Run("validate the records returned match the loaded records...", func(t *testing.T) {
		table, err := Load(testCSV, "test", "d6")
		if err != nil {
			t.Errorf("unexpected error, %s", err)
		}

		got := table.Records()

		if !reflect.DeepEqual(testCSV, got) {
			t.Errorf("want %v, got %v", testCSV, got)
		}
	})
}

func TestTable_RandomRow(t *testing.T) {
	t.Run("validate the random row returned...", func(t *testing.T) {
		table, err := Load(testCSV, "test", "d6")
		if err != nil {
			t.Errorf("unexpected error, %s", err)
		}

		got, i, err := table.RandomRow()
		if err != nil {
			t.Errorf("unexpected error getting random row, %s", err)
		}
		want := table.Rows[i-1].Results
		//rows 1, 4, and 6 have roll expressions, so I need to account for it
		if (i == 1) || (i == 4) || (i == 6) {
			want = []string{table.Rows[i-1].Results[0], rollString(table.Rows[i-1].Results[1]), table.Rows[i-1].Results[2]}
		}
		if !reflect.DeepEqual(want, got) {
			t.Errorf("want %v, got %v", want, got)
		}
	})

	t.Run("validate the random ranged row returned...", func(t *testing.T) {
		table, err := Load(rangedCSV, "test", "d6")
		if err != nil {
			t.Errorf("unexpected error, %s", err)
		}

		got, i, err := table.RandomRow()
		if err != nil {
			t.Errorf("unexpected error getting random row, %s", err)
		}
		want := table.Rows[0].Results
		switch i {
		case 1:
			want = table.Rows[0].Results
		case 2:
			want = table.Rows[0].Results
		case 3:
			want = table.Rows[1].Results
		case 4:
			want = table.Rows[1].Results
		case 5:
			want = table.Rows[2].Results
		case 6:
			want = table.Rows[2].Results
		}

		if !reflect.DeepEqual(want, got) {
			t.Errorf("want %v, got %v", want, got)
		}
	})
}

func TestTable_GetRow(t *testing.T) {
	t.Run("validate the row returned...", func(t *testing.T) {
		table, err := Load(testCSV, "test", "d6")
		if err != nil {
			t.Errorf("unexpected error, %s", err)
		}

		_, roll := dice.Roll(1, 6)
		got, err := table.GetRow(roll)
		if err != nil {
			t.Errorf("unexpected error getting row, %s", err)
		}
		want := table.Rows[roll-1].Results
		//rows 1, 4, and 6 have roll expressions, so I need to account for it
		if (roll == 1) || (roll == 4) || (roll == 6) {
			want = []string{table.Rows[roll-1].Results[0], rollString(table.Rows[roll-1].Results[1]), table.Rows[roll-1].Results[2]}
		}
		if !reflect.DeepEqual(want, got) {
			t.Errorf("want %v, got %v", want, got)
		}
	})

	t.Run("validate the ranged row returned...", func(t *testing.T) {
		table, err := Load(rangedCSV, "test", "d6")
		if err != nil {
			t.Errorf("unexpected error, %s", err)
		}
		_, roll := dice.Roll(1, 6)
		got, err := table.GetRow(roll)
		if err != nil {
			t.Errorf("unexpected error getting row, %s", err)
		}
		want := table.Rows[0].Results
		switch roll {
		case 1:
			want = table.Rows[0].Results
		case 2:
			want = table.Rows[0].Results
		case 3:
			want = table.Rows[1].Results
		case 4:
			want = table.Rows[1].Results
		case 5:
			want = table.Rows[2].Results
		case 6:
			want = table.Rows[2].Results
		}

		if !reflect.DeepEqual(want, got) {
			t.Errorf("want %v, got %v", want, got)
		}
	})

	t.Run("validate the ranged row returned if end of range requested...", func(t *testing.T) {
		table, err := Load(rangedWithExpressionCSV, "test", "d6")
		if err != nil {
			t.Errorf("unexpected error, %s", err)
		}
		got, err := table.GetRow(4)
		if err != nil {
			t.Errorf("unexpected error getting row, %s", err)
		}
		want := []string{table.Rows[1].Results[0], rollString(table.Rows[1].Results[1])}

		if !reflect.DeepEqual(want, got) {
			t.Errorf("want %v, got %v", want, got)
		}
	})
}

func TestTable_Expression(t *testing.T) {
	t.Run("validate the random row returned...", func(t *testing.T) {
		table, err := Load(testCSV, "test", "d6")
		if err != nil {
			t.Errorf("unexpected error, %s", err)
		}

		rows, err := table.Expression("2?test")
		if err != nil {
			t.Errorf("unexpected error getting random row, %s", err)
		}

		found := 0
		for _, got := range rows {
			for _, temp := range table.Rows {
				want := []string{temp.Results[0], rollString(temp.Results[1]), temp.Results[2]}
				if reflect.DeepEqual(want, got) {
					found++
					break
				}
			}
		}
		if found != 2 {
			t.Errorf("want 2, got %d", found)
		}
	})

	t.Run("validate the random ranged row returned...", func(t *testing.T) {
		table, err := Load(rangedCSV, "test", "d6")
		if err != nil {
			t.Errorf("unexpected error, %s", err)
		}

		rows, err := table.Expression("2?test")
		if err != nil {
			t.Errorf("unexpected error getting random row, %s", err)
		}

		found := 0
		for _, got := range rows {
			for _, temp := range table.Rows {
				want := temp.Results
				if reflect.DeepEqual(want, got) {
					found++
					break
				}
			}
		}
		if found != 2 {
			t.Errorf("want 2, got %d", found)
		}
	})

	t.Run("validate the specific row returned...", func(t *testing.T) {
		table, err := Load(testCSV, "test", "d6")
		if err != nil {
			t.Errorf("unexpected error, %s", err)
		}

		rows, err := table.Expression("3#test")
		if err != nil {
			t.Errorf("unexpected error getting specific row, %s", err)
		}

		if len(rows) == 2 { //header is included
			got := rows[1]
			want := table.Rows[2].Results
			if !reflect.DeepEqual(want, got) {
				t.Errorf("want %v, got %v", want, got)
			}
		} else {
			t.Errorf("want 2, got %d", len(rows))
		}
	})

	t.Run("validate the specific ranged row returned...", func(t *testing.T) {
		table, err := Load(rangedCSV, "test", "d6")
		if err != nil {
			t.Errorf("unexpected error, %s", err)
		}

		rows, err := table.Expression("4#test")
		if err != nil {
			t.Errorf("unexpected error getting specific row, %s", err)
		}

		if len(rows) == 2 { //header is included
			got := rows[1]
			want := table.Rows[1].Results
			if !reflect.DeepEqual(want, got) {
				t.Errorf("want %v, got %v", want, got)
			}
		} else {
			t.Errorf("want 2, got %d", len(rows))
		}

	})

	t.Run("validate the random row returned are unique...", func(t *testing.T) {
		table, err := Load(testCSV, "test", "d6")
		if err != nil {
			t.Errorf("unexpected error, %s", err)
		}

		rows, err := table.Expression("uni:2?test")
		if err != nil {
			t.Errorf("unexpected error getting random row, %s", err)
		}

		found := 0

		for _, got := range rows {
			for _, temp := range table.Rows {
				want := []string{temp.Results[0], rollString(temp.Results[1]), temp.Results[2]}
				if reflect.DeepEqual(want, got) {
					found++
					break
				}
			}
		}
		if found != 2 {
			t.Errorf("want 2, got %d", found)
		}

		if reflect.DeepEqual(rows[0], rows[1]) {
			t.Errorf("expected rows to be unique, got %v == %v", rows[0], rows[1])
		}
	})
}

func Test_RollableString(t *testing.T) {
	testCases := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "validate true is returned if value contains a roll expression...",
			value: "This string can be roll {{2d3+4}} times.",
			want:  true,
		},
		{
			name:  "validate true is returned if value contains roll expressions...",
			value: "This string can be roll {{2d3+4}} times. Or {{1d6}} times.",
			want:  true,
		},
		{
			name:  "validate false is returned if value does not contain any roll expressions...",
			value: "This string can not be rolled at all.",
			want:  false,
		},
		{
			name:  "validate false is returned if value does not contain any roll expressions in {{}}...",
			value: "This string can be roll 2d3+4 times.",
			want:  false,
		},
		{
			name:  "validate false is returned if value does not contain any valid roll expressions...",
			value: "This string can be roll {{2f3+a}} times.",
			want:  false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			got := RollableString(test.value)

			if got != test.want {
				t.Errorf("want %t, got %t", test.want, got)
			}
		})
	}
}

func Test_RangedRoll(t *testing.T) {
	testCases := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "validate true is returned if value is a range...",
			value: "1-4",
			want:  true,
		},
		{
			name:  "validate false is returned if value has too many dashes...",
			value: "1-4-8-9",
			want:  false,
		},
		{
			name:  "validate false is returned if value has non-numerics in first place...",
			value: "A-4",
			want:  false,
		},
		{
			name:  "validate false is returned if value has non-numerics in second place...",
			value: "6-B",
			want:  false,
		},
		{
			name:  "validate false is returned if value has non-numerics in both places...",
			value: "A-B",
			want:  false,
		},
		{
			name:  "validate false is returned if value is not a range...",
			value: "8",
			want:  false,
		},
		{
			name:  "validate false is returned if value is invalid...",
			value: "Not even close",
			want:  false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			got := RangedRoll(test.value)

			if got != test.want {
				t.Errorf("want %t, got %t", test.want, got)
			}
		})
	}
}

func Test_RangeInRoll(t *testing.T) {
	testCases := []struct {
		name      string
		roll      int
		rollRange string
		want      bool
	}{
		{
			name:      "validate true is returned if roll is in range...",
			roll:      3,
			rollRange: "1-4",
			want:      true,
		},
		{
			name:      "validate true is returned if roll is at start...",
			roll:      6,
			rollRange: "6-10",
			want:      true,
		},
		{
			name:      "validate true is returned if roll is at end...",
			roll:      8,
			rollRange: "2-8",
			want:      true,
		},
		{
			name:      "validate false is returned if roll is not in range...",
			roll:      8,
			rollRange: "1-4",
			want:      false,
		},
		{
			name:      "validate false is returned if roll range is not a range...",
			roll:      8,
			rollRange: "8",
			want:      false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			got := RollInRange(test.roll, test.rollRange)

			if got != test.want {
				t.Errorf("want %t, got %t", test.want, got)
			}
		})
	}
}