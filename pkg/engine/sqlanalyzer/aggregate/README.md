# Kwil Aggregate Restrictions

The following are the rules for aggregate queries in Kwil:

- An aggregate query is any query that contains an aggregate function
- Aggregate queries cannot contain columns in their result set UNLESS the columns are (1 of the following must be true):
    1. encapsulated AS THE FIRST ARGUMENT in an aggregate function in the result set
    2. included in the group by predicate
- The GROUP BY predicate expressions can only contain column expressions
- The HAVING predicate can only reference columns contained within the GROUP BY predicate
- Aggregate functions cannot contain subqueries
