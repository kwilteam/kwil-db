# Kwil Default Ordering

To guarantee determinism, Kwil has a default ordering that can be used for selects.  The default ordering rules are as follows:

- Each primary key column FOR EACH TABLE JOINED is ordered in ascending order
- Columns from all used tables are ordered alphabetically (first by table name, then by column name)
- Primary keys are given precedence alphabetically (e.g. column "age" will be ordered before column "name")
- User provided ordering is given precedence over default ordering
- If the user orders a primary key column, it will override the default ordering for that column
- If the query is a compound select, then all of the returned terms are ordered, instead of the primary keys.
The returned terms are ordered in the order that they are passed
