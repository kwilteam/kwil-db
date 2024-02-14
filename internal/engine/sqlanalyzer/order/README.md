# Ordering

To enforce guaranteed ordering, we enforce default ordering rules. Statements are broken down into their `*tree.SelectStmt` structs,
including nested statements (in the case of CTEs or subqueries).

Each `*tree.SelectStmt` is categorized as either being simple or compound. A compound is a SELECT that uses a compound operator,
such as UNION. A simple SELECT is a select that does not have any compound operator.

## Simple SELECTs

Simple SELECTs have the following rules applied:

- Each primary key column FOR EACH TABLE JOINED is ordered in ascending order. Table aliases will be ordered instead of name, if used.
- Primary keys from all used tables are ordered alphabetically (first by table name, then by column name) Primary keys are given precedence alphabetically (e.g. column "age" will be ordered before column "name")
- User provided ordering is given precedence over default ordering, and will therefore appear first in the statement.

If a simple SELECT has a GROUP BY clause, none of these rules apply, and instead it will simply order by all columns includes in the group by.

## Compound SELECTs

Compound SELECTs will be ordered by each returned column, in the order they appear. For the time being, compound selects with group bys will be rejected.
This is a remnant of a restriction imposed by SQLite's relatively rudimentary referencing, where it cannot order table aliases in a compound select.
If we wish to support GROUP BYs in compound selects in the future, we will need to dig deeper on Postgres's referencing. Since group bys in compounds are not
very common statements, we have decided to not support them for now.
