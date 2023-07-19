package order

import (
	"fmt"
	"sort"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

/*
	This file implements the default ordering of the Kwil spec.

	It modifies select statements to generate SQL that contains Kwil's default ordering if no ordering is provided.

	Kwil's default ordering rules are as follows:
	- Each primary key column FOR EACH TABLE JOINED is ordered in ascending order
	- Columns from all used tables are ordered alphabetically (first by table name, then by column name)
	- Primary keys are given precedence alphabetically (e.g. column "age" will be ordered before column "name")
	- User provided ordering is given precedence over default ordering
	- If the user orders a primary key column, it will override the default ordering for that column

	* From the SQLite docs (https://www.sqlite.org/lang_select.html#orderby)):

	If a SELECT statement that returns more than one row does not have an ORDER BY clause, the order in which the
	rows are returned is undefined. Or, if a SELECT statement does have an ORDER BY clause, then the list of
	expressions attached to the ORDER BY determine the order in which rows are returned to the user.

	In a compound SELECT statement, only the last or right-most simple SELECT may have an ORDER BY clause.
	That ORDER BY clause will apply across all elements of the compound. If the right-most element of a
	compound SELECT is a VALUES clause, then no ORDER BY clause is allowed on that statement.

	Rows are first sorted based on the results of evaluating the left-most expression in the ORDER BY list,
	then ties are broken by evaluating the second left-most expression and so on. The order in which two rows for
	which all ORDER BY expressions evaluate to equal values are returned is undefined. Each ORDER BY expression
	may be optionally followed by one of the keywords ASC (smaller values are returned first) or DESC (larger values
	are returned first). If neither ASC or DESC are specified, rows are sorted in ascending (smaller values
	first) order by default.

	SQLite considers NULL values to be smaller than any other values for sorting purposes. Hence, NULLs naturally
	appear at the beginning of an ASC order-by and at the end of a DESC order-by. This can be changed using the "ASC
	NULLS LAST" or "DESC NULLS FIRST" syntax.

	Each ORDER BY expression is processed as follows:

	-	If the ORDER BY expression is a constant integer K then the expression is considered an alias for the K-th
	column of the result set (columns are numbered from left to right starting with 1).

	- 	If the ORDER BY expression is an identifier that corresponds to the alias of one of the output columns, then
	the expression is considered an alias for that column.

	-	Otherwise, if the ORDER BY expression is any other expression, it is evaluated and the returned value used
	to order the output rows. If the SELECT statement is a simple SELECT, then an ORDER BY may contain any arbitrary
	expressions. However, if the SELECT is a compound SELECT, then ORDER BY expressions that are not aliases to output
	columns must be exactly the same as an expression used as an output column. // TODO: this is a big issue.  see notes from last night for potential solutions.

	For the purposes of sorting rows, values are compared in the same way as for comparison expressions. The collation
	sequence used to compare two text values is determined as follows:

	-	If the ORDER BY expression is assigned a collation sequence using the postfix COLLATE operator, then the specified
	collation sequence is used.

	-	Otherwise, if the ORDER BY expression is an alias to an expression that has been assigned a collation sequence
	using the postfix COLLATE operator, then the collation sequence assigned to the aliased expression is used.

	-	Otherwise, if the ORDER BY expression is a column or an alias of an expression that is a column, then the default
	collation sequence for the column is used.

	-	Otherwise, the BINARY collation sequence is used.

	In a compound SELECT statement, all ORDER BY expressions are handled as aliases for one of the result columns of the
	compound. If an ORDER BY expression is not an integer alias, then SQLite searches the left-most SELECT in the compound
	for a result column that matches either the second or third rules above. If a match is found, the search stops and the
	expression is handled as an alias for the result column that it has been matched against. Otherwise, the next SELECT
	to the right is tried, and so on. If no matching expression can be found in the result columns of any constituent SELECT,
	it is an error. Each term of the ORDER BY clause is processed separately and may be matched against result columns from
	different SELECT statements in the compound.
*/

type orderContext struct {
	MainTable    *types.Table
	JoinedTables []*types.Table
}

func (o *orderContext) generateOrder(term *orderableTerm) (*tree.OrderingTerm, error) {
	tbl, err := o.GetTable(term.Table)
	if err != nil {
		return nil, fmt.Errorf("failed to default generate order from term: %w", err)
	}

	return &tree.OrderingTerm{
		Expression: &tree.ExpressionColumn{
			Table:  tbl.Name,
			Column: term.Column,
		},
		OrderType:    tree.OrderTypeAsc,
		NullOrdering: tree.NullOrderingTypeLast, // I don't think this is needed, but just in case
	}, nil
}

func (o *orderContext) GetTable(name string) (*types.Table, error) {
	if o.MainTable.Name == name {
		return o.MainTable, nil
	}

	for _, tbl := range o.JoinedTables {
		if tbl.Name == name {
			return tbl, nil
		}
	}

	return nil, fmt.Errorf("table %s not found", name)
}

func (o *orderContext) requiredOrderingTerms() ([]*orderableTerm, error) {
	orderingTerms := []*orderableTerm{}
	orderedTables := orderTables(o.JoinedTables)
	for _, tbl := range orderedTables {
		required, err := getRequiredOrderingTerms(tbl)
		if err != nil {
			return nil, err
		}

		orderingTerms = append(orderingTerms, required...)
	}

	return orderingTerms, nil
}

func orderTables(tables []*types.Table) []*types.Table {
	sort.Slice(tables, func(i, j int) bool {
		return tables[i].Name < tables[j].Name
	})

	return tables
}

func getRequiredOrderingTerms(tbl *types.Table) ([]*orderableTerm, error) {
	pks, err := tbl.GetPrimaryKey()
	if err != nil {
		return nil, err
	}
	sort.Strings(pks)

	orderingTerms := []*orderableTerm{}
	for _, pk := range pks {
		orderingTerms = append(orderingTerms, &orderableTerm{
			Table:  tbl.Name,
			Column: pk,
		})
	}

	return orderingTerms, nil
}
