package table

import (
	"github.com/spf13/cobra"
)

func createTableCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create",
		Short: "Create is used for creating a new table.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}

// func CreateTable() {
// 	// ask the user if they would like to create the table raw, or with the table builder
// 	input, err := utils.PromptStringArr("Would you like to create the table yourself, or with the table builder?", []string{"Raw", "Table Builder"})
// 	if err != nil {
// 		color.Red("Error: %s", err)
// 		return
// 	}

// 	switch input {
// 	case "raw":
// 		promptCreateTable()
// 	case "table builder":
// 		launchTableBuilder()
// 	}
// }

// func promptCreateTable() {
// 	// first prompt the user to enter the table config
// 	str, err := utils.PromptStringInput("create table")
// 	if err != nil {
// 		fmt.Printf("Prompt failed %v\n", err)
// 		return
// 	}

// 	// set to lower case
// 	str = strings.ToLower(str)

// 	/*
// 		the string should follow the following format:
// 		<name> <column_name>:<column_type>:<constraint> <column_name>:<column_type>:<constraint> ...
// 	*/

// 	// first split the string by spaces
// 	split := strings.Split(str, " ")

// 	// the first element should be the name
// 	// now we need to parse the columns
// 	// iterate over 1: of the split
// 	var columns []Column
// 	for _, c := range split[1:] {
// 		// split the column by :
// 		s := strings.Split(c, ":")

// 		// some might not contain a constraint
// 		if len(s) == 2 {
// 			// no constraint
// 			columns = append(columns, Column{
// 				Name: s[0],
// 				Type: s[1],
// 			})
// 		} else if len(s) == 3 {
// 			// constraint
// 			columns = append(columns, Column{
// 				Name:       s[0],
// 				Type:       s[1],
// 				Constraint: s[2],
// 			})
// 		} else {
// 			color.Red("Invalid column format")
// 			fmt.Println("Column format should be <column_name>:<column_type>:<constraint> or just <column_name>:<column_type>")
// 			return
// 		}
// 	}

// 	// now we have the name and the columns
// 	// we can create the table

// 	table := NewTable{
// 		Name:    split[0],
// 		Columns: columns,
// 	}

// 	table.Check()
// 	table.GenerateDefaultQueries()
// }

// func launchTableBuilder() {
// 	// now we prompt the user to enter the name of the table
// 	fmt.Println("Please choose a name for your table")
// 	tableName, err := utils.PromptStringInput("Table name")
// 	if err != nil {
// 		fmt.Printf("Prompt failed %v\n", err)
// 		return
// 	}

// 	// check if the table name contains a space
// 	if containsSpace(tableName) {
// 		color.Red("Table name cannot contain a space")
// 		return
// 	}

// 	table := newTable(tableName)

// 	// now we ask the user what they would like to do next
// 	fmt.Println("Please choose what action you would like to perform.\n(Tables need at least one column.  Finishing without one will cancel table creation)")
// 	input, err := utils.PromptStringArr("Options", []string{"Add column", "Finish"})
// 	if err != nil {
// 		fmt.Printf("Prompt failed %v\n", err)
// 		return
// 	}

// 	switch input {
// 	case "add column":
// 		stop := false
// 		for !stop {
// 			table.AddColumn()
// 			input, err := utils.PromptStringArr("What would you like to do next?", []string{"Add column", "Finish"})
// 			if err != nil {
// 				fmt.Printf("Prompt failed %v\n", err)
// 				return
// 			}

// 			if input == "finish" {
// 				stop = true
// 				table.FinishTable()
// 			}
// 		}
// 	case "finish":
// 		table.FinishTable()
// 	}
// }

// func newTable(n string) *NewTable {
// 	return &NewTable{
// 		Name: n,
// 	}
// }

// type NewTable struct {
// 	Name    string
// 	Columns []Column
// }

// type Column struct {
// 	Name       string
// 	Type       string
// 	Constraint string
// }

// func (t *NewTable) AddColumn() {
// 	// now we prompt the user to enter the name of the column
// 	fmt.Println("Please choose a name for your column")
// 	columnName, err := utils.PromptStringInput("Column name")
// 	if err != nil {
// 		fmt.Printf("Prompt failed %v\n", err)
// 		return
// 	}

// 	// check if the column name contains a space
// 	if containsSpace(columnName) {
// 		color.Red("Column name cannot contain a space")
// 		return
// 	}

// 	// now we prompt the user to select the type of the column
// 	fmt.Println("Please select a type for your column")
// 	columnType, err := utils.PromptStringArr("Select column type", []string{"String", "Int32", "Int64", "Date", "DateTime", "Boolean"})
// 	if err != nil {
// 		fmt.Printf("Prompt failed %v\n", err)
// 		return
// 	}

// 	var ml uint64

// 	if columnType == "string" {
// 		// prompt for max length
// 		fmt.Println("Please enter the maximum length of the string (maximum allowed is 1024)")
// 		l, err := utils.PromptStringInput("Max length")
// 		if err != nil {
// 			fmt.Printf("Prompt failed %v\n", err)
// 			return
// 		}

// 		// convert to int16
// 		ml, err = strconv.ParseUint(l, 10, 16)
// 		if err != nil {
// 			color.Red("Prompt failed.  Please enter a valid integer")
// 			return
// 		}

// 		if ml > 1024 {
// 			color.Red("Max length cannot exceed 1024")
// 			return
// 		}
// 	}

// 	// now we prompt the user to select the constraint of the column
// 	fmt.Println("Please select a constraint for your column")
// 	columnConstraint, err := utils.PromptStringArr("Constraints", []string{"None", "PrimaryKey", "Unique", "NotNull", "Default"})
// 	if err != nil {
// 		fmt.Printf("Prompt failed %v\n", err)
// 		return
// 	}

// 	var cType string
// 	if columnType == "string" {
// 		cType = fmt.Sprintf("string(%d)", ml)
// 	} else {
// 		cType = columnType
// 	}

// 	// make the new column
// 	c := newColumn(columnName, cType, columnConstraint)

// 	fmt.Println()
// 	color.Green("New column:")
// 	c.PrintColumn()

// 	// ask if they want to add the column
// 	input, err := utils.PromptStringArr("Would you like to add this column?", []string{"Yes", "No"})
// 	if err != nil {
// 		fmt.Printf("Prompt failed %v\n", err)
// 		return
// 	}

// 	if input == "yes" {
// 		t.Columns = append(t.Columns, *c)
// 	}
// }

// func (t *NewTable) FinishTable() {
// 	// this sends the table to the database
// 	c := t.Check()
// 	if !c {
// 		color.Red("Aborting...")
// 		return
// 	}
// 	t.PrintTable()

// 	t.GenerateDefaultQueries()
// }

// func (t *NewTable) PrintTable() {
// 	color.Set(color.FgGreen)
// 	fmt.Println("Table name:", t.Name)
// 	color.Unset()
// 	for _, c := range t.Columns {
// 		fmt.Println("Column name:", c.Name)
// 		fmt.Println("Column type:", c.Type)
// 		fmt.Println("Column constraint:", c.Constraint)
// 		fmt.Println()
// 	}
// }

// func (c *Column) PrintColumn() {
// 	fmt.Println("Column name:", c.Name)
// 	fmt.Println("Column type:", c.Type)
// 	fmt.Println("Column constraint:", c.Constraint)
// 	fmt.Println()
// }

// func newColumn(n, t, c string) *Column {
// 	return &Column{
// 		Name:       n,
// 		Type:       t,
// 		Constraint: c,
// 	}
// }

// func containsSpace(s string) bool {
// 	for _, c := range s {
// 		if c == ' ' {
// 			return true
// 		}
// 	}

// 	return false
// }

// // Check performs a variety of checks on the table
// func (t *NewTable) Check() bool {
// 	// first ensure that there is one and only one primary key
// 	containsPK := false

// 	for _, c := range t.Columns {
// 		if c.Constraint == "primarykey" {
// 			if containsPK {
// 				color.Red("Table cannot contain more than one primary key")
// 				return false
// 			}
// 			containsPK = true
// 		}
// 	}

// 	if !containsPK {
// 		color.Red("Table must contain a primary key")
// 		return false
// 	}

// 	return true
// }

// // Generates the default insert, update, and delete queries for the table
// func (t *NewTable) GenerateDefaultQueries() {

// 	insertQ := Query{
// 		Name:       fmt.Sprintf("insert_%s", t.Name),
// 		Type:       "insert",
// 		Parameters: t.Columns,
// 	}

// 	updateQ := Query{
// 		Name:       fmt.Sprintf("update_%s", t.Name),
// 		Type:       "update",
// 		Parameters: t.Columns,
// 	}

// 	deleteQ := Query{
// 		Name: fmt.Sprintf("delete_%s", t.Name),
// 		Type: "delete",
// 	}

// 	// loop through columns to find the primary key and add it do the delete paramaters
// 	for _, c := range t.Columns {
// 		if c.Constraint == "primarykey" {
// 			deleteQ.Parameters = append(deleteQ.Parameters, c)
// 			deleteQ.Where = c.Name

// 			// add primary key to update where
// 			updateQ.Where = c.Name
// 		}
// 	}

// 	fmt.Println("The following queries have been generated:")
// 	insertQ.Print()
// 	updateQ.Print()
// 	deleteQ.Print()

// }

// type Query struct {
// 	Name       string
// 	Type       string
// 	Where      string
// 	Parameters []Column
// }

// func (q *Query) Print() {
// 	color.Set(color.Bold)
// 	color.Set(color.FgGreen)
// 	fmt.Println("Name:", q.Name)
// 	color.Unset()
// 	fmt.Println("	Type:", q.Type)
// 	fmt.Println("	Where:", q.Where)
// 	fmt.Printf("	Inputs:")
// 	for _, p := range q.Parameters {
// 		fmt.Printf(" %s:%s", p.Name, p.Type)
// 	}
// 	fmt.Println()
// }
