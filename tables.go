package tables

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
