package tables

import (
	"reflect"
	"testing"
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

var badCSV = [][]string{
	{"D3", "Result", "Description"},
	{"A", "Fight {{1d1}} rats", "The party runs across some dirty rats."},
	{"2:8", "No encounter", "Nothing to see here."},
	{"3", "A wolf can be heard nearby", "If the party is careful they may avoid the wolf."},
}

func Test_Load(t *testing.T) {
	table, err := Load(testCSV, "test", "d6")
	if err != nil {
		t.Errorf("unexpected error, %s", err)
	}

	meta := Meta{Name: "test", Headers: []string{"D6", "Result", "Description"}, ColumnCount: 3, RollableTable: true, RollExpression: "d6"}
	if !reflect.DeepEqual(meta, table.Meta) {
		t.Errorf("want %v, got %v", meta, table.Meta)
	}

	testCases := []struct {
		name  string
		index int
		want  Row
	}{
		{
			name:  "validate row 1...",
			index: 0,
			want:  Row{DieRoll: 1, RollRange: "", HasRollExpression: true, Results: []string{"1", "Fight {{1d1}} rats", "The party runs across some dirty rats."}},
		},
		{
			name:  "validate row 2...",
			index: 1,
			want:  Row{DieRoll: 2, RollRange: "", HasRollExpression: false, Results: []string{"2", "No encounter", "Nothing to see here."}},
		},
		{
			name:  "validate row 3...",
			index: 2,
			want:  Row{DieRoll: 3, RollRange: "", HasRollExpression: false, Results: []string{"3", "A wolf can be heard nearby", "If the party is careful they may avoid the wolf."}},
		},
		{
			name:  "validate row 4...",
			index: 3,
			want:  Row{DieRoll: 4, RollRange: "", HasRollExpression: true, Results: []string{"4", "{{1d1+1}} bats attack", "Angry bats swarm and attack the party."}},
		},
		{
			name:  "validate row 5...",
			index: 4,
			want:  Row{DieRoll: 5, RollRange: "", HasRollExpression: false, Results: []string{"5", "I can see you, can you see me?", "A whisper can be heard in the trees."}},
		},
		{
			name:  "validate row 6...",
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
	table, err := Load(rangedCSV, "ranged", "d6")
	if err != nil {
		t.Errorf("unexpected error, %s", err)
	}

	meta := Meta{Name: "ranged", Headers: []string{"D6", "Result"}, ColumnCount: 2, RollableTable: true, RollExpression: "d6"}
	if !reflect.DeepEqual(meta, table.Meta) {
		t.Errorf("want %v, got %v", meta, table.Meta)
	}

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
	table, err := Load(badCSV, "bad", "d6")
	if err == nil {
		t.Error("expected error, got nil")
	}

	want := Table{}

	if !reflect.DeepEqual(want, table) {
		t.Errorf("want %v, got %v", want, table)
	}
}

func Test_rangedRoll(t *testing.T) {
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
			got := rangedRoll(test.value)

			if got != test.want {
				t.Errorf("want %t, got %t", test.want, got)
			}
		})
	}
}
