package tables

//Type indicates the type of table.
type Type string

const (
	//Simple type tables have columns  without any roll expressions
	Simple Type = "Simple Table"
	//Advanced type tables have some columns with roll expressions, this means they will need to be evaluated
	Advanced Type = "Advanced Table"
)

//Table represents a table with meta data and rows
type Table struct {
	Meta Meta  `json:"meta"`
	Rows []Row `json:"rows"`
}

//Meta stores metadata for a table
type Meta struct {
	Type           Type     `json:"type"`
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
