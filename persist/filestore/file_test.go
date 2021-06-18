package filestore

import (
	"os"
	"reflect"
	"testing"

	"github.com/fantastical-world/tables"
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

var testJSON = `{"testfile":{"meta":{"name":"testfile","title":"","flavor_text":"","campaign":"","headers":["D6","Result","Description"],"column_count":3,"rollable_table":true,"roll_expression":"d6"},"rows":[{"die_roll":1,"roll_range":"","has_roll_expression":true,"results":["1","Fight {{1d1}} rats","The party runs across some dirty rats."]},{"die_roll":2,"roll_range":"","has_roll_expression":false,"results":["2","No encounter","Nothing to see here."]},{"die_roll":3,"roll_range":"","has_roll_expression":false,"results":["3","A wolf can be heard nearby","If the party is careful they may avoid the wolf."]},{"die_roll":4,"roll_range":"","has_roll_expression":true,"results":["4","{{1d1+1}} bats attack","Angry bats swarm and attack the party."]},{"die_roll":5,"roll_range":"","has_roll_expression":false,"results":["5","I can see you, can you see me?","A whisper can be heard in the trees."]},{"die_roll":6,"roll_range":"","has_roll_expression":true,"results":["6","A pile of bones covers {{1d1}}GP","You found some loot."]}]}}`
var badJSON = `{"counter": "dracula"}`

func Test_New(t *testing.T) {
	_ = New("./test.json")
}

func TestDatabase_ExistingFile(t *testing.T) {
	file, err := os.Create("./test.json")
	if err != nil {
		t.Errorf("unexpected error creating file, %s", err)
	}
	_, err = file.WriteString(testJSON)
	if err != nil {
		t.Errorf("unexpected error writing json, %s", err)
	}

	err = file.Close()
	if err != nil {
		t.Errorf("unexpected error closing file, %s", err)
	}

	db := New("./test.json")
	table, err := tables.Load(testCSV, "testfile", "d6")
	if err != nil {
		t.Errorf("unexpected error loading table, %s", err)
	}

	t.Run("validate that get table correctly returns table...", func(t *testing.T) {
		got, err := db.GetTable("testfile")
		if err != nil {
			t.Errorf("unexpected error getting table, %s", err)
		}

		if !reflect.DeepEqual(table, got) {
			t.Errorf("want %v, got %v", table, got)
		}
	})

	err = os.Remove("./test.json")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestDatabase_SaveTable(t *testing.T) {
	db := New("./test.json")
	table, err := tables.Load(testCSV, "test", "d6")
	if err != nil {
		t.Errorf("unexpected error loading table, %s", err)
	}

	t.Run("validate that save table correctly persists table...", func(t *testing.T) {
		err := db.SaveTable(table)
		if err != nil {
			t.Errorf("unexpected error saving table, %s", err)
		}

		got, err := db.GetTable("test")
		if err != nil {
			t.Errorf("unexpected error getting table, %s", err)
		}

		if !reflect.DeepEqual(table, got) {
			t.Errorf("want %v, got %v", table, got)
		}
	})

	t.Run("validate that save table returns an error if table invalid...", func(t *testing.T) {
		emptyTable := tables.Table{}
		err := db.SaveTable(emptyTable)
		if err == nil {
			t.Error("expected an error, but none encountered")
		}
	})

	err = os.Remove("./test.json")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestDatabase_GetTable(t *testing.T) {
	db := New("./test.json")
	table, err := tables.Load(testCSV, "test", "d6")
	if err != nil {
		t.Errorf("unexpected error loading table, %s", err)
	}

	t.Run("validate that get table correctly returns table...", func(t *testing.T) {
		err := db.SaveTable(table)
		if err != nil {
			t.Errorf("unexpected error saving table, %s", err)
		}

		got, err := db.GetTable("test")
		if err != nil {
			t.Errorf("unexpected error getting table, %s", err)
		}

		if !reflect.DeepEqual(table, got) {
			t.Errorf("want %v, got %v", table, got)
		}
	})

	t.Run("validate that get table returns an error if table does not exist...", func(t *testing.T) {
		got, err := db.GetTable("IDONTEXIST")
		if err != tables.ErrTableDoesNotExist {
			t.Errorf("want %s, got %s", tables.ErrTableDoesNotExist, err)
		}
		want := tables.Table{}
		if !reflect.DeepEqual(want, got) {
			t.Errorf("want %v, got %v", want, got)
		}
	})

	err = os.Remove("./test.json")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestDatabase_DeleteTable(t *testing.T) {
	db := New("./test.json")
	table, err := tables.Load(testCSV, "test", "d6")
	if err != nil {
		t.Errorf("unexpected error loading table, %s", err)
	}

	t.Run("validate that delete table removes table...", func(t *testing.T) {
		err := db.SaveTable(table)
		if err != nil {
			t.Errorf("unexpected error saving table, %s", err)
		}

		got, err := db.GetTable("test")
		if err != nil {
			t.Errorf("unexpected error getting table, %s", err)
		}

		if !reflect.DeepEqual(table, got) {
			t.Errorf("want %v, got %v", table, got)
		}

		err = db.DeleteTable("test")
		if err != nil {
			t.Errorf("unexpected error deleting table, %s", err)
		}

		got, err = db.GetTable("test")
		if err != tables.ErrTableDoesNotExist {
			t.Errorf("want %s, got %s", tables.ErrTableDoesNotExist, err)
		}
		want := tables.Table{}
		if !reflect.DeepEqual(want, got) {
			t.Errorf("want %v, got %v", want, got)
		}
	})

	t.Run("validate that delete table returns an error if table does not exist...", func(t *testing.T) {
		err = db.DeleteTable("IDONTEXIST")
		if err != tables.ErrTableDoesNotExist {
			t.Errorf("want %s, got %s", tables.ErrTableDoesNotExist, err)
		}
	})

	err = os.Remove("./test.json")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestDatabase_ListTables(t *testing.T) {
	db := New("./test.json")
	tableTest, err := tables.Load(testCSV, "test", "")
	if err != nil {
		t.Errorf("unexpected error loading table, %s", err)
	}

	tableAnother, err := tables.Load(testCSV, "another", "d6")
	if err != nil {
		t.Errorf("unexpected error loading table, %s", err)
	}

	tableRanged, err := tables.Load(testCSV, "ranged", "d6")
	if err != nil {
		t.Errorf("unexpected error loading table, %s", err)
	}

	t.Run("validate table listing...", func(t *testing.T) {
		err := db.SaveTable(tableTest)
		if err != nil {
			t.Errorf("unexpected error saving table, %s", err)
		}

		err = db.SaveTable(tableAnother)
		if err != nil {
			t.Errorf("unexpected error saving table, %s", err)
		}

		err = db.SaveTable(tableRanged)
		if err != nil {
			t.Errorf("unexpected error saving table, %s", err)
		}

		got, err := db.ListTables()
		if err != nil {
			t.Errorf("unexpected error listing tables, %s", err)
		}

		want := []string{"another,d6,true", "ranged,d6,true", "test,,false"}
		if !reflect.DeepEqual(want, got) {
			t.Errorf("want %v, got %v", want, got)
		}
	})

	err = os.Remove("./test.json")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestDatabase_BadFile(t *testing.T) {
	file, err := os.Create("./test.json")
	if err != nil {
		t.Errorf("unexpected error creating file, %s", err)
	}
	_, err = file.WriteString(badJSON)
	if err != nil {
		t.Errorf("unexpected error writing json, %s", err)
	}

	err = file.Close()
	if err != nil {
		t.Errorf("unexpected error closing file, %s", err)
	}

	db := New("./test.json") //this should continue with empty tables
	table, err := tables.Load(testCSV, "test", "d6")
	if err != nil {
		t.Errorf("unexpected error loading table, %s", err)
	}

	t.Run("validate that saving table works even if file had bad json...", func(t *testing.T) {
		err := db.SaveTable(table) //will recreate file
		if err != nil {
			t.Errorf("unexpected error saving table, %s", err)
		}

		got, err := db.GetTable("test")
		if err != nil {
			t.Errorf("unexpected error getting table, %s", err)
		}

		if !reflect.DeepEqual(table, got) {
			t.Errorf("want %v, got %v", table, got)
		}
	})

	err = os.Remove("./test.json")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestDatabase_BadFileStore(t *testing.T) {
	file, err := os.Create("./test.json")
	if err != nil {
		t.Errorf("unexpected error creating file, %s", err)
	}
	_, err = file.WriteString(testJSON)
	if err != nil {
		t.Errorf("unexpected error writing json, %s", err)
	}

	file.Chmod(0444)

	err = file.Close()
	if err != nil {
		t.Errorf("unexpected error closing file, %s", err)
	}

	db := New("./test.json")
	table, err := tables.Load(testCSV, "test", "d6")
	if err != nil {
		t.Errorf("unexpected error loading table, %s", err)
	}

	t.Run("validate that save table returns an error if file can not beused...", func(t *testing.T) {
		err := db.SaveTable(table)
		if err == nil {
			t.Error("expected error, err was nil")
		}
	})

	t.Run("validate that delete table returns an error if file can not beused...", func(t *testing.T) {
		err := db.DeleteTable("testfile")
		if err == nil {
			t.Error("expected error, err was nil")
		}
	})

	err = os.Remove("./test.json")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}
