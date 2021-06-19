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

var testJSON = `{"meta":{"name":"testfile","title":"","flavor_text":"","campaign":"","headers":["D6","Result","Description"],"column_count":3,"rollable_table":true,"roll_expression":"d6"},"rows":[{"die_roll":1,"roll_range":"","has_roll_expression":true,"results":["1","Fight {{1d1}} rats","The party runs across some dirty rats."]},{"die_roll":2,"roll_range":"","has_roll_expression":false,"results":["2","No encounter","Nothing to see here."]},{"die_roll":3,"roll_range":"","has_roll_expression":false,"results":["3","A wolf can be heard nearby","If the party is careful they may avoid the wolf."]},{"die_roll":4,"roll_range":"","has_roll_expression":true,"results":["4","{{1d1+1}} bats attack","Angry bats swarm and attack the party."]},{"die_roll":5,"roll_range":"","has_roll_expression":false,"results":["5","I can see you, can you see me?","A whisper can be heard in the trees."]},{"die_roll":6,"roll_range":"","has_roll_expression":true,"results":["6","A pile of bones covers {{1d1}}GP","You found some loot."]}]}`
var badJSON = `{"counter": "dracula"}`

func Test_New(t *testing.T) {
	t.Run("validate that new with valid directory location is successful...", func(t *testing.T) {
		_, err := New(".")
		if err != nil {
			t.Errorf("unexpected error, %s", err)
		}
	})

	t.Run("validate that new with invalid directory location returns an error...", func(t *testing.T) {
		_, err := New("//safsdfsfsdf///sdfsdfs/")
		if err == nil {
			t.Error("expected error, err was nil")
		}
	})
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

	db, err := New(".")
	if err != nil {
		t.Errorf("unexpected error, %s", err)
	}
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
	db, err := New(".")
	if err != nil {
		t.Errorf("unexpected error, %s", err)
	}
	table, err := tables.Load(testCSV, "test", "d6")
	if err != nil {
		t.Errorf("unexpected error loading table, %s", err)
	}

	t.Run("validate that save table correctly persists table...", func(t *testing.T) {
		err := db.SaveTable(table)
		if err != nil {
			t.Errorf("unexpected error saving table, %s", err)
		}

		if err == nil {
			err = os.Remove(table.Hash() + ".json")
			if err != nil {
				t.Errorf("unexpected err encountered deleting file, %s", err)
			}
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
}

func TestDatabase_GetTable(t *testing.T) {
	db := FileStore{tables: make(map[string]tables.Table)}
	table, err := tables.Load(testCSV, "test", "d6")
	if err != nil {
		t.Errorf("unexpected error loading table, %s", err)
	}
	db.tables["test"] = table

	t.Run("validate that get table correctly returns table...", func(t *testing.T) {
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
}

func TestDatabase_DeleteTable(t *testing.T) {
	db, err := New(".")
	if err != nil {
		t.Errorf("unexpected error, %s", err)
	}
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
}

func TestDatabase_ListTables(t *testing.T) {
	db := FileStore{tables: make(map[string]tables.Table)}
	tableTest, err := tables.Load(testCSV, "test", "")
	if err != nil {
		t.Errorf("unexpected error loading table, %s", err)
	}
	db.tables["test"] = tableTest

	tableAnother, err := tables.Load(testCSV, "another", "d6")
	if err != nil {
		t.Errorf("unexpected error loading table, %s", err)
	}
	db.tables["another"] = tableAnother

	tableRanged, err := tables.Load(testCSV, "ranged", "d6")
	if err != nil {
		t.Errorf("unexpected error loading table, %s", err)
	}
	db.tables["ranged"] = tableRanged

	t.Run("validate table listing...", func(t *testing.T) {
		got, err := db.ListTables()
		if err != nil {
			t.Errorf("unexpected error listing tables, %s", err)
		}

		want := []string{"another,d6,true", "ranged,d6,true", "test,,false"}
		if !reflect.DeepEqual(want, got) {
			t.Errorf("want %v, got %v", want, got)
		}
	})
}

func TestDatabase_BadFile(t *testing.T) {
	db, err := New(".")
	if err != nil {
		t.Errorf("unexpected error, %s", err)
	}
	table, err := tables.Load(testCSV, "noluck", "d6")
	if err != nil {
		t.Errorf("unexpected error loading table, %s", err)
	}

	file, err := os.Create(table.Hash() + ".json")
	if err != nil {
		t.Errorf("unexpected error creating file, %s", err)
	}

	file.Chmod(0444)

	t.Run("validate that saving table returns error if file can not be written to...", func(t *testing.T) {
		err := db.SaveTable(table)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	err = os.Remove(table.Hash() + ".json")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

/*
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

	db, err := New(".")
	if err != nil {
		t.Errorf("unexpected error, %s", err)
	}
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
*/
