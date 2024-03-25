# Kwil SQL AST

This package contains the specification for the Kwil Engine Statement Interface.  It includes information regarding supported functionalities, limitations / restrictions, etc, for Kwil SQL Engine Statements.  This document outlines these statements using Golang structs and interfaces.  For more information on structs and interfaces, read: [structs](<https://golangbot.com/structs/>) / [interfaces](<https://golangbot.com/interfaces-part-1/#:~:text=In%20Go%2C%20an%20interface%20is,how%20to%20implement%20these%20methods.>)

**Most users should not use this interface, but instead use the SQL interface.**  This interface is meant to provide context on all possible functionalities of Kwil, as well as provide an interface for tooling to be built around.

## Overview

There are 4 main structs that are meant to be used as "entry points" / "outermost interfaces" for instructions to the Kwil Database Engine.  These structs with transpile to standalone instructions to the SQL engine, and will gracefully handle errors:

- Insert
- Update
- Delete
- Select

Other structs can be used from the interface, but they will likely not function as standalone statements, and do not have graceful error handling in the case of being used incorrectly.  It is recommended for users to only use these to embed within one of the four "outermost" interfaces.

### Insert

Insert is the outermost struct for containing any stand-alone Insert statement.

```go
type Insert struct {
    CTE        []*CTE
    InsertStmt *InsertStmt
}
```

### Update

Update is the outermost struct for containing any stand-alone Update statement.

```go
type Update struct {
    CTE        []*CTE
    UpdateStmt *UpdateStmt
}
```

### Delete

Delete is the outermost struct for containing any stand-alone Delete statement.

```go
type Delete struct {
    CTE        []*CTE
    DeleteStmt *DeleteStmt
}
```

### Select

Select is the outermost struct for containing any stand-alone Select statement.

```go
type Select struct {
    CTE        []*CTE
    SelectStmt *SelectStmt
}
```

## Other Interfaces

The rest of the structs / interfaces / enumerators supported by Kwil are listed below, alphebetically.

#### CollationType

Collations are used for specifying the bit pattern that should be used for characters.

```go
type CollationType string

const (
    CollationTypeBinary CollationType = "BINARY"
    CollationTypeNoCase CollationType = "NOCASE"
    CollationTypeRTrim  CollationType = "RTRIM"
)
```

_In SQL:_

```sql
SELECT *
FROM users
WHERE username = 'sAtoSHi'
# Start of collation
COLLATE NOCASE;
# End of collation
```

#### Conflict Target

A conflict target is used to specify the what indexed column(s) should be watched in the case of a UNIQUE conflict.

```go
type ConflictTarget struct {
    IndexedColumns []string
    Where          Expression
}
```

_In SQL:_

```sql
INSERT INTO users (username, age, user_type)
VALUES ('satoshi', '42', 'registered')
# Start of conflict target
ON CONFLICT (username)
    WHERE user_type = 'unregistered'
    ## End of conflict target
    DO UPDATE SET age='42', user_type = 'registered'
    WHERE username = 'satoshi';
```

#### CTE

Common table expressions (CTE) are used to create temporary tables / views that can be used later in a statement.

```go
type CTE struct {
    Table   string
    Columns []string
    Select  *SelectStmt
}
```

_In SQL:_

```sql
# Start of CTE
WITH users_followers AS (
    SELECT followed_id AS followers
    FROM followers
    WHERE follower_id = (
        SELECT id FROM users
        WHERE username = 'satoshi'
    )
)
# End of CTE
SELECT *
FROM posts
WHERE author_id IN (
    SELECT followers
    FROM users_followers
)
ORDER BY post_height DESC;
```

#### DeleteStmt

A delete statement is a full statement that specifies a record deletion.  Unlike the pure 'Delete' struct, it does not contain common table expressions.

```go
type DeleteStmt struct {
    QualifiedTableName *QualifiedTableName
    Where              Expression
    Returning          *ReturningClause
}
```

_In SQL:_

```sql
# Start of DeleteStmt
DELETE FROM users
WHERE username = 'sbf'
RETURNING (
    SELECT *
    FROM users
    LIMIT 10
);
# End of DeleteStmt
```

#### Expression

Expressions are combinations of one or more values, identifiers, comparisons, and other expressions that evaluate to a value / set of values.  There are multiple types of expressions in Kwil, so they have [their own page](./expressions.md).

#### Functions

Functions are invokable clauses that take inputs and return a value.  Functions are valid expressions, and take other expressions as inputs.  A full list of functions supported by Kwil can be found on the [Functions page](./functions.md).

#### GroupBy

The GROUP BY clause is used to group rows that have the same values in specified columns into aggregated data.

GROUP BY clauses can also contain a HAVING clause, which filters the returned results based on some condition.

```go
type GroupBy struct {
    Expressions []Expression
    Having      Expression
}
```

_In SQL:_

```sql
SELECT u.username, COUNT(f.follower_id)
FROM users AS u
LEFT JOIN followers AS f
    ON u.id = f.followed_id
# Start of GROUP BY
GROUP BY f.followed_id, u.id
HAVING COUNT(f.follower_id) > 100;
# End of GROUP BY
```

_The above query retrieves the usernames and follower count for all users who have more than 100 followers_

#### InsertStmt

The InsertStmt is the content of an Insert occurring after any common table expressions.

```go
type InsertStmt struct {
    InsertType      InsertType
    Table           string
    TableAlias      string
    Columns         []string
    Values          [][]Expression
    Upsert          *Upsert
    ReturningClause *ReturningClause
}
```

_In SQL:_

**(TODO: I am quite certain this is not a valid query)**

```sql
INSERT INTO users (username, age, wallet_address)
VALUES ($username, $age, @caller)
ON CONFLICT (username)
    WHERE wallet_address = @caller
    DO UPDATE SET
        age = $age
        WHERE username = $username
RETURNING (
    SELECT username
    FROM users
    WHERE age = $age
    LIMIT 10
);
```

_The above query inserts a new user into a table, owned by the wallet address calling the query.  If the username is already registered and owned by the caller, it will update the age to the new value.  The query then returns up to 10 other users who have the same age._

#### InsertType

The Kwil interface contains different InsertType enumerations to specify different Insert operations.  The default is 'InsertTypeInsert'

```go
type InsertType uint8

const (
    InsertTypeInsert InsertType = iota
    InsertTypeReplace
    InsertTypeInsertOrReplace
)
```

_In SQL:_

```sql
# Begin InsertType
INSERT OR REPLACE /*End InsertType*/ INTO users (username, age)
VALUES ('satoshi', 42)
```

#### Relation
A Relation is used to specify the table(s) that should be used in a SELECT statement.
It can be a table, a subquery, or joined tables.

#### RelationTable

A RelationTable is used to specify a table that should be used in a SELECT statement.

```go
type RelationTable struct {
    Name  string
    Alias string
}

```

_In SQL:_

```sql
SELECT *
# Start of RelationTable
FROM users AS u
# End of RelationTable
WHERE username = 'satoshi';
```

#### RelationSubquery

A RelationSubquery is used to specify a subquery that should be used in a SELECT statement.

```go
type RelationSubquery struct {
    Select *SelectStmt
    Alias  string
}
```

_In SQL:_

```sql
SELECT *
# Start of RelationSubquery
FROM (
    SELECT *
    FROM users
    WHERE age > 18
) AS u
# End of RelationSubquery
WHERE username = 'satoshi';
```

#### RelationJoin

A RelationJoin is used to specify joined tables that should be used in a SELECT statement.

```go
type RelationJoin struct {
    Relation        Relation
    Joins           []*JoinPredicate
}
```

_In SQL:_

```sql
SELECT p.title, p.content
# Start of RelationJoin
FROM posts AS p
INNER JOIN users as u
    ON p.author_id = u.id
# End of JoinClause
```

#### JoinPredicate

A join predicate contains the type of join, the subject / target table of the join, and the ON predicate (a.k.a. 'Contraint') on which the tables should be joined.

```go
type JoinPredicate struct {
    JoinOperator *JoinOperator
    Table        Relation
    Constraint   Expression
}
```

_In SQL:_

```sql
SELECT p.title, p.content
FROM posts AS p
# Start of JoinPredicate
INNER JOIN users as u
    ON p.author_id = u.id
# End of JoinPredicate
```

#### JoinOperator

The JoinOperator specifies the type of join that should be performed.  It takes a join type, and has the option to make them "natural" and "outer" joins.  Inner joins and regular joins cannot be made "outer".

```go
type JoinOperator struct {
    Natural  bool
    JoinType JoinType
    Outer    bool
}
```

_In SQL:_

```sql
SELECT p.title, p.content
FROM posts AS p
/*Start of JoinOperator*/OUTER LEFT JOIN /*End of JoinOperator*/ users as u
    ON p.author_id = u.id
```

#### JoinType

The JoinType is an enumerator that specifies the type of join to be performed.

```go
type JoinType uint8

const (
    JoinTypeJoin JoinType = iota
    JoinTypeInner
    JoinTypeLeft
    JoinTypeRight
    JoinTypeFull
)
```

_In SQL:_

```sql
SELECT p.title, p.content
FROM posts AS p
OUTER /*Start of JoinType*/ LEFT JOIN /*End of JoinType*/ users as u
    ON p.author_id = u.id
```

#### Limit

A LIMIT clause is used to specify record limits.  It takes an expression to limit (usually a literal, but can be anything).  It optionally takes an offset as well, or a second expression.

```go
type Limit struct {
    Expression       Expression
    Offset           Expression
    SecondExpression Expression
}
```

_In SQL:_

```sql
SELECT *
FROM users
# Limit clause:
LIMIT 10 OFFSET 20
```

#### OrderBy

An ORDER BY clause specifies the way in which query results should be ordered.  A single ORDER BY can contain multiple ordering terms.

```go
type OrderBy struct {
    OrderingTerms []*OrderingTerm
}
```

_In SQL:_

```sql
SELECT *
FROM users
## OrderBy with 2 ordering terms
ORDER BY username COLLATE BINARY,
age ASC NULLS LAST
```

#### OrderingTerm

An ordering term is a singular elements of an ORDER BY clause.  It contains an expression to order on, and an optional collation, order type (ASC/DESC), and terms for null ordering.

```go
type OrderingTerm struct {
    Expression   Expression
    Collation    CollationType
    OrderType    OrderType
    NullOrdering NullOrderingType
}
```

_In SQL:_

```sql
SELECT *
FROM users
## ordering term 1
ORDER BY username COLLATE BINARY,
## ordering term 2
age ASC NULLS LAST
```

#### OrderType

The OrderType is used to specify the order in which nuemeric values should appear.

```go
type OrderType string

const (
    OrderTypeNone OrderType = ""
    OrderTypeAsc  OrderType = "ASC"
    OrderTypeDesc OrderType = "DESC"
)
```

#### NullOrderingType

The NullOrderingType specifies the order in which nulls should be returned in a query.

```go
type NullOrderingType string

const (
    NullOrderingTypeNone  NullOrderingType = ""
    NullOrderingTypeFirst NullOrderingType = "NULLS FIRST"
    NullOrderingTypeLast  NullOrderingType = "NULLS LAST"
)
```

#### QualifiedTableName

A QualifiedTableName (QTN) is used to specify update and delete target tables.  Unlike Insert and Select (which do not except QTNs), QTNs in Update and Delete can optionally contain instructions for whether or not indexes should apply.

```go
type QualifiedTableName struct {
    TableName  string
    TableAlias string
    IndexedBy  string
    NotIndexed bool
}
```

_In SQL:_

```sql
DELETE FROM
# Start of qualified table name
comments INDEXED BY comment_index
# End of qualified table name
WHERE post_id = 5;
```

#### Returning Clause

A returning clause is used to specify data that should be returned from a query that is not a SELECT.

```go
type ReturningClause struct {
    Returned []*ReturningClauseColumn
}
```

_In SQL:_

```sql
INSERT INTO posts (id, title, content, author_id)
VALUES ($id, $title, $content, $author_id)
# Returning Clause
RETURNING id, title, content
```

#### ReturningClauseColumn

A returning clause column is used to specify what column should be returned in a returning clause.

```go
type ReturningClauseColumn struct {
    All        bool
    Expression Expression
    Alias      string
}
```

_In SQL:_

```sql
INSERT INTO posts (id, title, content, author_id)
VALUES ($id, $title, $content, $author_id)
# id, title, content are all returning clause columns:
RETURNING id, title, content
```

#### SelectStmt

The SelectStmt is the a select statement without common table expressions.

```go
type SelectStmt struct {
    SelectCore *SelectCore
    OrderBy    *OrderBy
    Limit      *Limit
}
```

_In SQL:_

```sql
SELECT *
FROM users
ORDER BY username
LIMIT 10
```

#### SelectCore

The SelectCore is the select statement without OrderBy and Limit.  SelectCore's can be compounded on each other using CompoundOperators.

```go
type SelectCore struct {
    SelectType SelectType
    Columns    []string
    From       *FromClause
    Where      Expression
    GroupBy    *GroupBy
    Compound   *CompoundOperator
}
```

_In SQL:_

```sql
SELECT u.username, COUNT(f.follower_id)
FROM users AS u
LEFT JOIN followers AS f
    ON u.id = f.followed_id
GROUP BY f.followed_id, u.id
HAVING COUNT(f.follower_id) > 100;
```

#### SelectType

SelectType is an enumerator that specifies the type of SELECT that should be performed.  The default type is SelectTypeAll.

```go
type SelectType uint8

const (
    SelectTypeAll SelectType = iota
    SelectTypeDistinct
)
```

#### FromClause

The FROM clause is a wrapper around Relation.  The differentiation is necessary to make future features easier to
implement.

```go
type FromClause struct {
    Relation Relation
}
```

_In SQL:_

```sql
SELECT u.username, COUNT(f.follower_id)
# Start FromClause
FROM users AS u
LEFT JOIN followers AS f
    ON u.id = f.followed_id
# End FromClause
GROUP BY f.followed_id, u.id
HAVING COUNT(f.follower_id) > 100;
```

#### CompoundOperator

A compound operator is used to combine two SelectCores into a single set.

```go
type CompoundOperator struct {
    Operator     CompoundOperatorType
    SelectClause *SelectCore
}
```

_In SQL:_

```sql
SELECT *
FROM posts
WHERE author_id = 1
# Begin compound operator
UNION ALL
SELECT *
FROM posts
WHERE author_id = 2
# End compound operator
```

_This is not the best way to write this query, but it displays how CompoundOperator can be used_

#### CompoundOperatorType

The compound operator type is an enumerator specifiyng the type of compound operator that should be used.

```go
type CompoundOperatorType uint8

const (
    CompoundOperatorTypeUnion CompoundOperatorType = iota
    CompoundOperatorTypeUnionAll
    CompoundOperatorTypeIntersect
    CompoundOperatorTypeExcept
)
```

#### Relation

The Relation is used to specify a value that can be either a table or a subquery or a join clause.  There are several
different types,
which are displayed below:

```go
type RelationTable struct {
    Name  string
    Alias string
}

type RelationSubquery struct {
    Select *SelectStmt
    Alias  string
}

type RelationJoin struct {
    Relation Relation
    Joins    []*JoinPredicate
}
```

_In SQL:_

```sql
SELECT u.username, COUNT(f.follower_id)
# Start
FROM users AS u
# End
LEFT JOIN followers AS f
    ON u.id = f.followed_id
GROUP BY f.followed_id, u.id
HAVING COUNT(f.follower_id) > 100;
```

#### UpdateStmt

The UpdateStmt is an Update without common table expressions.

```go
type UpdateStmt struct {
    Or                 UpdateOr
    QualifiedTableName *QualifiedTableName
    UpdateSetClause    []*UpdateSetClause
    From               *FromClause
    Where              Expression
    Returning          *ReturningClause
}
```

_In SQL:_

```sql
UPDATE users
SET age = 25
WHERE username = 'satoshi'
```

#### UpdateOr

The UpdateOr clause is used to specify if there should be an alternative for the OR statement.

```go
type UpdateOr string

const (
    UpdateOrAbort    UpdateOr = "ABORT"
    UpdateOrFail     UpdateOr = "FAIL"
    UpdateOrIgnore   UpdateOr = "IGNORE"
    UpdateOrReplace  UpdateOr = "REPLACE"
    UpdateOrRollback UpdateOr = "ROLLBACK"
)
```

#### UpdateSetClause

The UpdateSetClause is used to specify what column(s) should be used in an update statement.

```go
type UpdateSetClause struct {
    Columns    []string
    Expression Expression
}
```

_In SQL:_

```sql
UPDATE users
# Begin singular UpdateSetClause
SET age = 25
# End UpdateSetClause
WHERE username = 'satoshi'
```

#### Upsert

An Upsert clause is used to specify actions that should occur when a UNIQUE conflict occurs.

```go
type Upsert struct {
    ConflictTarget *ConflictTarget
    Type           UpsertType
    Updates        []*UpdateSetClause
    Where          Expression
}
```

_In SQL:_

```sql
INSERT INTO users (username, age, user_type)
VALUES ('satoshi', '42', 'registered')
# Start of conflict target
ON CONFLICT (username)
    WHERE user_type = 'unregistered'
    ## End of conflict target
    DO UPDATE SET age='42', user_type = 'registered'
    WHERE username = 'satoshi';
```

#### UpsertType

The UpsertType is used to specify what should occur on an upsert:

```go
type UpsertType uint8

const (
    UpsertTypeDoNothing UpsertType = iota
    UpsertTypeDoUpdate
)
```
