package kvstore

import (
	"os"
	"testing"
)

func TestNewDatabase(t *testing.T) {
	_, err := New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}
}

func TestLoadTable(t *testing.T) {
	db, err := New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test", "d6")
	if err != nil {
		t.Errorf("error loading table was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test2", "")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestGetTable(t *testing.T) {
	db, err := New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test", "d6")
	if err != nil {
		t.Errorf("error while loading table was not expected, but err was encountered %s\n", err)
	}

	results, err := db.GetTable("test")
	if err != nil {
		t.Errorf("error while getting table was not expected, but err was encountered %s\n", err)
	}
	count := 7 //header + 6 records
	if len(results) != count {
		t.Errorf("expected %d records, actual records %d", count, len(results))
	}

	for i, row := range results {
		if len(row) != 3 {
			t.Errorf("expected %d fields, actual fields %d", 3, len(row))
		}
		switch i {
		case 0:
			if (row[0] != "D6") || (row[1] != "Result") || (row[2] != "Description") {
				t.Errorf("expected heard to contain [D6][Result][Description], actual values [%s][%s][%s]", row[0], row[1], row[2])
			}
		case 1:
			if (row[0] != "1") || (row[1] != "Fight {{1d1}} rats") || (row[2] != "The party runs across some dirty rats.") {
				t.Errorf("expected heard to contain [1][Fight {{1d1}} rats][The party runs across some dirty rats.], actual values [%s][%s][%s]", row[0], row[1], row[2])
			}
		case 2:
			if (row[0] != "2") || (row[1] != "No encounter") || (row[2] != "Nothing to see here.") {
				t.Errorf("expected heard to contain [2][No encounter][Nothing to see here.], actual values [%s][%s][%s]", row[0], row[1], row[2])
			}
		case 3:
			if (row[0] != "3") || (row[1] != "A wolf can be heard nearby") || (row[2] != "If the party is careful they may avoid the wolf.") {
				t.Errorf("expected heard to contain [3][A wolf can be heard nearby][If the party is careful they may avoid the wolf.], actual values [%s][%s][%s]", row[0], row[1], row[2])
			}
		case 4:
			if (row[0] != "4") || (row[1] != "{{1d1+1}} bats attack") || (row[2] != "Angry bats swarm and attack the party.") {
				t.Errorf("expected heard to contain [4][{{1d1+1}} bats attack][Angry bats swarm and attack the party.], actual values [%s][%s][%s]", row[0], row[1], row[2])
			}
		case 5:
			if (row[0] != "5") || (row[1] != "I can see you, can you see me?") || (row[2] != "A whisper can be heard in the trees.") {
				t.Errorf("expected heard to contain [5][I can see you, can you see me?][A whisper can be heard in the trees.], actual values [%s][%s][%s]", row[0], row[1], row[2])
			}
		case 6:
			if (row[0] != "6") || (row[1] != "A pile of bones covers {{1d1}}GP") || (row[2] != "You found some loot.") {
				t.Errorf("expected heard to contain [6][A pile of bones covers {{1d1}}GP][You found some loot.], actual values [%s][%s][%s]", row[0], row[1], row[2])
			}
		}
	}

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestTableExpression(t *testing.T) {
	db, err := New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test", "d6")
	if err != nil {
		t.Errorf("error while loading table was not expected, but err was encountered %s\n", err)
	}

	results, err := db.TableExpression("2?test")
	if err != nil {
		t.Errorf("error with table expression was not expected, but err was encountered %s\n", err)
	}
	count := 3 //header + 2 random rows
	if len(results) != count {
		t.Errorf("expected %d rows, actual rows %d", count, len(results))
	}

	if (results[0][0] != "D6") || (results[0][1] != "Result") || (results[0][2] != "Description") {
		t.Errorf("expected heard to contain [D6][Result][Description], actual values [%s][%s][%s]", results[0][0], results[0][1], results[0][2])
	}

	for i, row := range results {
		if len(row) != 3 {
			t.Errorf("expected %d fields, actual fields %d", 3, len(row))
		}

		if i == 0 {
			continue
		}

		switch row[0] {
		case "1":
			if (row[0] != "1") || (row[1] != "Fight 1 rats") || (row[2] != "The party runs across some dirty rats.") {
				t.Errorf("expected heard to contain [1][Fight 1 rats][The party runs across some dirty rats.], actual values [%s][%s][%s]", row[0], row[1], row[2])
			}
		case "2":
			if (row[0] != "2") || (row[1] != "No encounter") || (row[2] != "Nothing to see here.") {
				t.Errorf("expected heard to contain [2][No encounter][Nothing to see here.], actual values [%s][%s][%s]", row[0], row[1], row[2])
			}
		case "3":
			if (row[0] != "3") || (row[1] != "A wolf can be heard nearby") || (row[2] != "If the party is careful they may avoid the wolf.") {
				t.Errorf("expected heard to contain [3][A wolf can be heard nearby][If the party is careful they may avoid the wolf.], actual values [%s][%s][%s]", row[0], row[1], row[2])
			}
		case "4":
			if (row[0] != "4") || (row[1] != "2 bats attack") || (row[2] != "Angry bats swarm and attack the party.") {
				t.Errorf("expected heard to contain [4][2 bats attack][Angry bats swarm and attack the party.], actual values [%s][%s][%s]", row[0], row[1], row[2])
			}
		case "5":
			if (row[0] != "5") || (row[1] != "I can see you, can you see me?") || (row[2] != "A whisper can be heard in the trees.") {
				t.Errorf("expected heard to contain [5][I can see you, can you see me?][A whisper can be heard in the trees.], actual values [%s][%s][%s]", row[0], row[1], row[2])
			}
		case "6":
			if (row[0] != "6") || (row[1] != "A pile of bones covers 1GP") || (row[2] != "You found some loot.") {
				t.Errorf("expected heard to contain [6][A pile of bones covers 1GP][You found some loot.], actual values [%s][%s][%s]", row[0], row[1], row[2])
			}
		}
	}

	results, err = db.TableExpression("3#test")
	if err != nil {
		t.Errorf("error with table expression was not expected, but err was encountered %s\n", err)
	}
	count = 2 //header + row for roll 3
	if len(results) != count {
		t.Errorf("expected %d rows, actual rows %d", count, len(results))
	}

	if (results[0][0] != "D6") || (results[0][1] != "Result") || (results[0][2] != "Description") {
		t.Errorf("expected heard to contain [D6][Result][Description], actual values [%s][%s][%s]", results[0][0], results[0][1], results[0][2])
	}
	if (results[1][0] != "3") || (results[1][1] != "A wolf can be heard nearby") || (results[1][2] != "If the party is careful they may avoid the wolf.") {
		t.Errorf("expected heard to contain [3][A wolf can be heard nearby][If the party is careful they may avoid the wolf.], actual values [%s][%s][%s]", results[1][0], results[1][1], results[1][2])
	}

	results, err = db.TableExpression("uni:6?test")
	if err != nil {
		t.Errorf("error with table expression was not expected, but err was encountered %s\n", err)
	}
	count = 7 //header + 6 unique rows
	if len(results) != count {
		t.Errorf("expected %d rows, actual rows %d", count, len(results))
	}

	var rolls []string
	for i, roll := range results {
		if i == 0 {
			continue
		}
		rolls = append(rolls, roll[0])
	}
	//since there are 6 rows we should see each one since we asked for unique rows
	if !contains(rolls, "1") {
		t.Errorf("expected to find [1] in rolls, but it was not found: %s\n", rolls)
	}
	if !contains(rolls, "2") {
		t.Errorf("expected to find [2] in rolls, but it was not found: %s\n", rolls)
	}
	if !contains(rolls, "3") {
		t.Errorf("expected to find [3] in rolls, but it was not found: %s\n", rolls)
	}
	if !contains(rolls, "4") {
		t.Errorf("expected to find [4] in rolls, but it was not found: %s\n", rolls)
	}
	if !contains(rolls, "5") {
		t.Errorf("expected to find [5] in rolls, but it was not found: %s\n", rolls)
	}
	if !contains(rolls, "6") {
		t.Errorf("expected to find [6] in rolls, but it was not found: %s\n", rolls)
	}

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestRandomRow(t *testing.T) {
	db, err := New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test", "d6")
	if err != nil {
		t.Errorf("error while loading table was not expected, but err was encountered %s\n", err)
	}

	results, roll, err := db.RandomRow("test")
	if err != nil {
		t.Errorf("error with random row was not expected, but err was encountered %s\n", err)
	}
	count := 3 //3 fields
	if len(results) != count {
		t.Errorf("expected %d fields, actual fields %d", count, len(results))
	}

	switch roll {
	case 1:
		if (results[0] != "1") || (results[1] != "Fight 1 rats") || (results[2] != "The party runs across some dirty rats.") {
			t.Errorf("expected heard to contain [1][Fight 1 rats][The party runs across some dirty rats.], actual values [%s][%s][%s]", results[0], results[1], results[2])
		}
	case 2:
		if (results[0] != "2") || (results[1] != "No encounter") || (results[2] != "Nothing to see here.") {
			t.Errorf("expected heard to contain [2][No encounter][Nothing to see here.], actual values [%s][%s][%s]", results[0], results[1], results[2])
		}
	case 3:
		if (results[0] != "3") || (results[1] != "A wolf can be heard nearby") || (results[2] != "If the party is careful they may avoid the wolf.") {
			t.Errorf("expected heard to contain [3][A wolf can be heard nearby][If the party is careful they may avoid the wolf.], actual values [%s][%s][%s]", results[0], results[1], results[2])
		}
	case 4:
		if (results[0] != "4") || (results[1] != "2 bats attack") || (results[2] != "Angry bats swarm and attack the party.") {
			t.Errorf("expected heard to contain [4][2 bats attack][Angry bats swarm and attack the party.], actual values [%s][%s][%s]", results[0], results[1], results[2])
		}
	case 5:
		if (results[0] != "5") || (results[1] != "I can see you, can you see me?") || (results[2] != "A whisper can be heard in the trees.") {
			t.Errorf("expected heard to contain [5][I can see you, can you see me?][A whisper can be heard in the trees.], actual values [%s][%s][%s]", results[0], results[1], results[2])
		}
	case 6:
		if (results[0] != "6") || (results[1] != "A pile of bones covers 1GP") || (results[2] != "You found some loot.") {
			t.Errorf("expected heard to contain [6][A pile of bones covers 1GP][You found some loot.], actual values [%s][%s][%s]", results[0], results[1], results[2])
		}
	}

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestGetRow(t *testing.T) {
	db, err := New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test", "d6")
	if err != nil {
		t.Errorf("error while loading table was not expected, but err was encountered %s\n", err)
	}

	results, err := db.GetRow(1, "test")
	if err != nil {
		t.Errorf("error with get row was not expected, but err was encountered %s\n", err)
	}
	count := 3 //3 fields
	if len(results) != count {
		t.Errorf("expected %d fields, actual fields %d", count, len(results))
	}
	if (results[0] != "1") || (results[1] != "Fight 1 rats") || (results[2] != "The party runs across some dirty rats.") {
		t.Errorf("expected heard to contain [1][Fight 1 rats][The party runs across some dirty rats.], actual values [%s][%s][%s]", results[0], results[1], results[2])
	}

	results, err = db.GetRow(2, "test")
	if err != nil {
		t.Errorf("error with get row was not expected, but err was encountered %s\n", err)
	}
	if len(results) != count {
		t.Errorf("expected %d fields, actual fields %d", count, len(results))
	}
	if (results[0] != "2") || (results[1] != "No encounter") || (results[2] != "Nothing to see here.") {
		t.Errorf("expected heard to contain [2][No encounter][Nothing to see here.], actual values [%s][%s][%s]", results[0], results[1], results[2])
	}

	results, err = db.GetRow(3, "test")
	if err != nil {
		t.Errorf("error with get row was not expected, but err was encountered %s\n", err)
	}
	if len(results) != count {
		t.Errorf("expected %d fields, actual fields %d", count, len(results))
	}
	if (results[0] != "3") || (results[1] != "A wolf can be heard nearby") || (results[2] != "If the party is careful they may avoid the wolf.") {
		t.Errorf("expected heard to contain [3][A wolf can be heard nearby][If the party is careful they may avoid the wolf.], actual values [%s][%s][%s]", results[0], results[1], results[2])
	}

	results, err = db.GetRow(4, "test")
	if err != nil {
		t.Errorf("error with get row was not expected, but err was encountered %s\n", err)
	}
	if len(results) != count {
		t.Errorf("expected %d fields, actual fields %d", count, len(results))
	}
	if (results[0] != "4") || (results[1] != "2 bats attack") || (results[2] != "Angry bats swarm and attack the party.") {
		t.Errorf("expected heard to contain [4][2 bats attack][Angry bats swarm and attack the party.], actual values [%s][%s][%s]", results[0], results[1], results[2])
	}

	results, err = db.GetRow(5, "test")
	if err != nil {
		t.Errorf("error with get row was not expected, but err was encountered %s\n", err)
	}
	if len(results) != count {
		t.Errorf("expected %d fields, actual fields %d", count, len(results))
	}
	if (results[0] != "5") || (results[1] != "I can see you, can you see me?") || (results[2] != "A whisper can be heard in the trees.") {
		t.Errorf("expected heard to contain [5][I can see you, can you see me?][A whisper can be heard in the trees.], actual values [%s][%s][%s]", results[0], results[1], results[2])
	}

	results, err = db.GetRow(6, "test")
	if err != nil {
		t.Errorf("error with get row was not expected, but err was encountered %s\n", err)
	}
	if len(results) != count {
		t.Errorf("expected %d fields, actual fields %d", count, len(results))
	}
	if (results[0] != "6") || (results[1] != "A pile of bones covers 1GP") || (results[2] != "You found some loot.") {
		t.Errorf("expected heard to contain [6][A pile of bones covers 1GP][You found some loot.], actual values [%s][%s][%s]", results[0], results[1], results[2])
	}

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestRangedRows(t *testing.T) {
	db, err := New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test-ranged.csv", "test", "d6")
	if err != nil {
		t.Errorf("error while loading table was not expected, but err was encountered %s\n", err)
	}

	results, roll, err := db.RandomRow("test")
	if err != nil {
		t.Errorf("error with random row was not expected, but err was encountered %s\n", err)
	}
	count := 2 //2 fields
	if len(results) != count {
		t.Errorf("expected %d fields, actual fields %d", count, len(results))
	}

	switch roll {
	case 1:
		if (results[0] != "1-2") || (results[1] != "You rolled a 1 or 2") {
			t.Errorf("expected heard to contain [1-2][You rolled a 1 or 2], actual values [%s][%s]", results[0], results[1])
		}
	case 2:
		if (results[0] != "1-2") || (results[1] != "You rolled a 1 or 2") {
			t.Errorf("expected heard to contain [1-2][You rolled a 1 or 2], actual values [%s][%s]", results[0], results[1])
		}
	case 3:
		if (results[0] != "3-4") || (results[1] != "You rolled a 3 or 4") {
			t.Errorf("expected heard to contain [3-4][You rolled a 3 or 4], actual values [%s][%s]", results[0], results[1])
		}
	case 4:
		if (results[0] != "3-4") || (results[1] != "You rolled a 3 or 4") {
			t.Errorf("expected heard to contain [3-4][You rolled a 3 or 4], actual values [%s][%s]", results[0], results[1])
		}
	case 5:
		if (results[0] != "5-6") || (results[1] != "You rolled a 5 or 6") {
			t.Errorf("expected heard to contain [5-6][You rolled a 5 or 6], actual values [%s][%s]", results[0], results[1])
		}
	case 6:
		if (results[0] != "5-6") || (results[1] != "You rolled a 5 or 6") {
			t.Errorf("expected heard to contain [5-6][You rolled a 5 or 6], actual values [%s][%s]", results[0], results[1])
		}
	}

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestGetHeader(t *testing.T) {
	db, err := New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test", "d6")
	if err != nil {
		t.Errorf("error while loading table was not expected, but err was encountered %s\n", err)
	}

	results, err := db.GetHeader("test")
	if err != nil {
		t.Errorf("error with get header was not expected, but err was encountered %s\n", err)
	}
	count := 3 //header + 2 random rows
	if len(results) != count {
		t.Errorf("expected %d rows, actual rows %d", count, len(results))
	}

	if (results[0] != "D6") || (results[1] != "Result") || (results[2] != "Description") {
		t.Errorf("expected heard to contain [D6][Result][Description], actual values [%s][%s][%s]", results[0], results[1], results[2])
	}

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestListTables(t *testing.T) {
	db, err := New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test", "d6")
	if err != nil {
		t.Errorf("error loading table was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test2", "")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	tables, err := db.ListTables()
	if err != nil {
		t.Errorf("error while listing tables was not expected, but err was encountered %s\n", err)
	}

	if !contains(tables, "test,d6,Advanced Table,true") {
		t.Errorf("expected to find [test,d6,Advanced Table,true] table, but it was not found: %s\n", tables)
	}

	if !contains(tables, "test2,,Advanced Table,false") {
		t.Errorf("expected to find [test2,,Advanced Table,false] table, but it was not found: %s\n", tables)
	}

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestWriteTable(t *testing.T) {
	db, err := New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test", "d6")
	if err != nil {
		t.Errorf("error loading table was not expected, but err was encountered %s\n", err)
	}

	err = db.WriteTable("test", "./write.csv")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}

	err = db.LoadTable("./write.csv", "write", "d6")
	if err != nil {
		t.Errorf("error loading table was not expected, but err was encountered %s\n", err)
	}

	results, err := db.GetRow(1, "write")
	if err != nil {
		t.Errorf("error with get row was not expected, but err was encountered %s\n", err)
	}
	count := 3 //3 fields
	if len(results) != count {
		t.Errorf("expected %d fields, actual fields %d", count, len(results))
	}
	if (results[0] != "1") || (results[1] != "Fight 1 rats") || (results[2] != "The party runs across some dirty rats.") {
		t.Errorf("expected heard to contain [1][Fight 1 rats][The party runs across some dirty rats.], actual values [%s][%s][%s]", results[0], results[1], results[2])
	}

	results, err = db.GetRow(2, "write")
	if err != nil {
		t.Errorf("error with get row was not expected, but err was encountered %s\n", err)
	}
	if len(results) != count {
		t.Errorf("expected %d fields, actual fields %d", count, len(results))
	}
	if (results[0] != "2") || (results[1] != "No encounter") || (results[2] != "Nothing to see here.") {
		t.Errorf("expected heard to contain [2][No encounter][Nothing to see here.], actual values [%s][%s][%s]", results[0], results[1], results[2])
	}

	results, err = db.GetRow(3, "write")
	if err != nil {
		t.Errorf("error with get row was not expected, but err was encountered %s\n", err)
	}
	if len(results) != count {
		t.Errorf("expected %d fields, actual fields %d", count, len(results))
	}
	if (results[0] != "3") || (results[1] != "A wolf can be heard nearby") || (results[2] != "If the party is careful they may avoid the wolf.") {
		t.Errorf("expected heard to contain [3][A wolf can be heard nearby][If the party is careful they may avoid the wolf.], actual values [%s][%s][%s]", results[0], results[1], results[2])
	}

	results, err = db.GetRow(4, "write")
	if err != nil {
		t.Errorf("error with get row was not expected, but err was encountered %s\n", err)
	}
	if len(results) != count {
		t.Errorf("expected %d fields, actual fields %d", count, len(results))
	}
	if (results[0] != "4") || (results[1] != "2 bats attack") || (results[2] != "Angry bats swarm and attack the party.") {
		t.Errorf("expected heard to contain [4][2 bats attack][Angry bats swarm and attack the party.], actual values [%s][%s][%s]", results[0], results[1], results[2])
	}

	results, err = db.GetRow(5, "write")
	if err != nil {
		t.Errorf("error with get row was not expected, but err was encountered %s\n", err)
	}
	if len(results) != count {
		t.Errorf("expected %d fields, actual fields %d", count, len(results))
	}
	if (results[0] != "5") || (results[1] != "I can see you, can you see me?") || (results[2] != "A whisper can be heard in the trees.") {
		t.Errorf("expected heard to contain [5][I can see you, can you see me?][A whisper can be heard in the trees.], actual values [%s][%s][%s]", results[0], results[1], results[2])
	}

	results, err = db.GetRow(6, "write")
	if err != nil {
		t.Errorf("error with get row was not expected, but err was encountered %s\n", err)
	}
	if len(results) != count {
		t.Errorf("expected %d fields, actual fields %d", count, len(results))
	}
	if (results[0] != "6") || (results[1] != "A pile of bones covers 1GP") || (results[2] != "You found some loot.") {
		t.Errorf("expected heard to contain [6][A pile of bones covers 1GP][You found some loot.], actual values [%s][%s][%s]", results[0], results[1], results[2])
	}

	err = os.Remove("./write.csv")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestDeleteTable(t *testing.T) {
	db, err := New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test", "d6")
	if err != nil {
		t.Errorf("error loading table was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test2", "")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	tables, err := db.ListTables()
	if err != nil {
		t.Errorf("error while listing tables was not expected, but err was encountered %s\n", err)
	}

	if !contains(tables, "test,d6,Advanced Table,true") {
		t.Errorf("expected to find [test,d6,Advanced Table,true] table, but it was not found: %s\n", tables)
	}

	if !contains(tables, "test2,,Advanced Table,false") {
		t.Errorf("expected to find [test2,,Advanced Table,false] table, but it was not found: %s\n", tables)
	}

	err = db.Delete("test2")
	if err != nil {
		t.Errorf("error while deleting tables was not expected, but err was encountered %s\n", err)
	}

	tables, err = db.ListTables()
	if err != nil {
		t.Errorf("error while listing tables was not expected, but err was encountered %s\n", err)
	}

	if !contains(tables, "test,d6,Advanced Table,true") {
		t.Errorf("expected to find [test,d6,Advanced Table,true] table, but it was not found: %s\n", tables)
	}

	if contains(tables, "test2,,Advanced Table,false") {
		t.Errorf("expected not to find [test2,,Advanced Table,false] table, but it was found: %s\n", tables)
	}

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestGetMeta(t *testing.T) {
	db, err := New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test", "d6")
	if err != nil {
		t.Errorf("error loading table was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test2", "")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	meta, err := db.GetMeta("test")
	if err != nil {
		t.Errorf("error getting table meta was not expected, but err was encountered %s\n", err)
	}

	if meta.Name != "test" {
		t.Errorf("expected meta name to be [test], actual [%s]\n", meta.Name)
	}
	if meta.RollExpression != "d6" {
		t.Errorf("expected meta roll expression to be [d6], actual [%s]\n", meta.RollExpression)
	}
	if meta.RollableTable != true {
		t.Errorf("expected meta rollable table to be [true], actual [%t]\n", meta.RollableTable)
	}

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestReplacingTables(t *testing.T) {
	db, err := New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test", "d6")
	if err != nil {
		t.Errorf("error while loading table was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test-ranged.csv", "test", "d6")
	if err != nil {
		t.Errorf("error while loading table was not expected, but err was encountered %s\n", err)
	}

	results, roll, err := db.RandomRow("test")
	if err != nil {
		t.Errorf("error with random row was not expected, but err was encountered %s\n", err)
	}
	count := 2 //2 fields
	if len(results) != count {
		t.Errorf("expected %d fields, actual fields %d", count, len(results))
	}

	switch roll {
	case 1:
		if (results[0] != "1-2") || (results[1] != "You rolled a 1 or 2") {
			t.Errorf("expected heard to contain [1-2][You rolled a 1 or 2], actual values [%s][%s]", results[0], results[1])
		}
	case 2:
		if (results[0] != "1-2") || (results[1] != "You rolled a 1 or 2") {
			t.Errorf("expected heard to contain [1-2][You rolled a 1 or 2], actual values [%s][%s]", results[0], results[1])
		}
	case 3:
		if (results[0] != "3-4") || (results[1] != "You rolled a 3 or 4") {
			t.Errorf("expected heard to contain [3-4][You rolled a 3 or 4], actual values [%s][%s]", results[0], results[1])
		}
	case 4:
		if (results[0] != "3-4") || (results[1] != "You rolled a 3 or 4") {
			t.Errorf("expected heard to contain [3-4][You rolled a 3 or 4], actual values [%s][%s]", results[0], results[1])
		}
	case 5:
		if (results[0] != "5-6") || (results[1] != "You rolled a 5 or 6") {
			t.Errorf("expected heard to contain [5-6][You rolled a 5 or 6], actual values [%s][%s]", results[0], results[1])
		}
	case 6:
		if (results[0] != "5-6") || (results[1] != "You rolled a 5 or 6") {
			t.Errorf("expected heard to contain [5-6][You rolled a 5 or 6], actual values [%s][%s]", results[0], results[1])
		}
	}

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestReplacingTablesStandard(t *testing.T) {
	db, err := New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	err = db.LoadTable("./test.csv", "test", "")
	if err != nil {
		t.Errorf("error while loading table was not expected, but err was encountered %s\n", err)
	}

	results, err := db.GetTable("test")
	if err != nil {
		t.Errorf("error while getting table was not expected, but err was encountered %s\n", err)
	}
	count := 7 //header + 6 rows
	if len(results) != count {
		t.Errorf("expected %d rows, actual rows %d", count, len(results))
	}

	err = db.LoadTable("./test-ranged.csv", "test", "")
	if err != nil {
		t.Errorf("error while loading table was not expected, but err was encountered %s\n", err)
	}

	results, err = db.GetTable("test")
	if err != nil {
		t.Errorf("error while getting table was not expected, but err was encountered %s\n", err)
	}
	count = 4 //header + 3 rows
	if len(results) != count {
		t.Errorf("expected %d rows, actual rows %d", count, len(results))
	}

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func TestErrors(t *testing.T) {
	db, err := New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}

	_, err = db.GetTable("nope")
	if err == nil {
		t.Errorf("expected an err, but none occured\n")
	} else {
		if err.Error() != "table [nope] does not exist" {
			t.Errorf("expected err [table [nope] does not exist], actual [%s]\n", err)
		}
	}

	err = db.LoadTable("./bad.csv", "bad", "d3")
	if err == nil {
		t.Errorf("expected an err, but none occured\n")
	} else {
		if err.Error() != "first column must be an integer since it represents a die roll" {
			t.Errorf("expected err [first column must be an integer since it represents a die roll], actual [%s]\n", err)
		}
	}

	_, err = db.TableExpression("heyoplayo")
	if err == nil {
		t.Errorf("expected an err, but none occured\n")
	} else {
		if err.Error() != "not a valid table expression, must be ?table or n?table or n#table (e.g. ?npc, 2?npc, 3#npc)" {
			t.Errorf("expected err [not a valid table expression, must be ?table or n?table or n#table (e.g. ?npc, 2?npc, 3#npc)], actual [%s]\n", err)
		}
	}

	_, err = db.TableExpression("0#bad")
	if err == nil {
		t.Errorf("expected an err, but none occured\n")
	} else {
		if err.Error() != "not a valid table expression, a request to show a specific row must include a row number" {
			t.Errorf("expected err [not a valid table expression, a request to show a specific row must include a row number], actual [%s]\n", err)
		}
	}

	err = db.LoadTable("./test.csv", "testnoroll", "")
	if err != nil {
		t.Errorf("error while loading table was not expected, but err was encountered %s\n", err)
	}

	_, err = db.TableExpression("0?testnoroll")
	if err == nil {
		t.Errorf("expected an err, but none occured\n")
	} else {
		if err.Error() != "not a rollable table, no roll expression available" {
			t.Errorf("expected err [not a rollable table, no roll expression available], actual [%s]\n", err)
		}
	}

	_, _, err = db.RandomRow("no")
	if err == nil {
		t.Errorf("expected an err, but none occured\n")
	} else {
		if err.Error() != "table [no] does not exist" {
			t.Errorf("expected err [table [no] does not exist], actual [%s]\n", err)
		}
	}

	_, err = db.GetRow(2, "not")
	if err == nil {
		t.Errorf("expected an err, but none occured\n")
	} else {
		if err.Error() != "table [not] does not exist" {
			t.Errorf("expected err [table [not] does not exist], actual [%s]\n", err)
		}
	}

	err = db.LoadTable("./test.csv", "testgood", "d6")
	if err != nil {
		t.Errorf("error while loading table was not expected, but err was encountered %s\n", err)
	}

	_, err = db.GetRow(8, "testgood")
	if err == nil {
		t.Errorf("expected an err, but none occured\n")
	} else {
		if err.Error() != "value for [8] does not exist" {
			t.Errorf("expected err [value for [8] does not exist], actual [%s]\n", err)
		}
	}

	_, err = db.GetHeader("nope")
	if err == nil {
		t.Errorf("expected an err, but none occured\n")
	} else {
		if err.Error() != "table [nope] does not exist" {
			t.Errorf("expected err [table [nope] does not exist], actual [%s]\n", err)
		}
	}

	_, err = db.GetMeta("nopes")
	if err == nil {
		t.Errorf("expected an err, but none occured\n")
	} else {
		if err.Error() != "table [nopes] does not exist" {
			t.Errorf("expected err [table [nopes] does not exist], actual [%s]\n", err)
		}
	}

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
