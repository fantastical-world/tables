package kvstore

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

func Test_New(t *testing.T) {
	_ = New("./test.db")
}

func TestDatabase_SaveTable(t *testing.T) {
	db := New("./test.db")
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

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestDatabase_GetTable(t *testing.T) {
	db := New("./test.db")
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

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestDatabase_DeleteTable(t *testing.T) {
	db := New("./test.db")
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

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestDatabase_ListTables(t *testing.T) {
	db := New("./test.db")
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

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func Test_BadDatabaseErrors(t *testing.T) {
	db := New("./bad.db")

	err := db.SaveTable(tables.Table{})
	if err == nil {
		t.Error("expected an error, but none encountered\n")
	}
	_, err = db.GetTable("test")
	if err == nil {
		t.Error("expected an error, but none encountered\n")
	}
	_, err = db.ListTables()
	if err == nil {
		t.Error("expected an error, but none encountered\n")
	}
	err = db.DeleteTable("test")
	if err == nil {
		t.Error("expected an error, but none encountered\n")
	}
}

func Test_MessupDatabaseErrors(t *testing.T) {
	db := New("./messedup.db")

	_, err := db.GetTable("notatable")
	if err == nil {
		t.Error("expected an error, but none encountered\n")
	}
	_, err = db.ListTables()
	if err == nil {
		t.Error("expected an error, but none encountered\n")
	}
}
