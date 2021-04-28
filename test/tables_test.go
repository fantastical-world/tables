package test

import (
	"os"
	"testing"

	"tables/database/kvstore"
)

func TestNewDatabase(t *testing.T) {
	_, err := kvstore.New("./test.db")
	if err != nil {
		t.Errorf("error was not expected, but err was encountered %s\n", err)
	}
}

func TestLoadTable(t *testing.T) {
	db, err := kvstore.New("./test.db")
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
	db, err := kvstore.New("./test.db")
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
	db, err := kvstore.New("./test.db")
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

	err = os.Remove("./test.db")
	if err != nil {
		t.Errorf("unexpected err encountered, %s", err)
	}
}
