# Expressions

Expressions are valuatable clauses that are comprised of literals, identifiers, comparisons, functions, subqueries, and other expressions. All expressions are interchangeable, however they will not all necessarily create valid statements.

#### ExpressionLiteral

An expression literal is used to define any concrete, static value.  These are things like strings, integers, and floating point numbers.  Floating point numbers will immediately be rounded when they are passed to the engine.

```go
type ExpressionLiteral struct {
    Value interface{}
}
```

_In SQL:_

```sql
INSERT INTO users (username, age)
VALUES (
    /*string literal:*/ 'satoshi',
    /*integer literal:*/ 42
);
```

#### ExpressionBindParameter

An expression bind parameter is used to denote user input.  User's can store statements containing bind parameters in their database engine, and later execute them, while passing in different values for the parameter.

They can also be used as a method to protect against SQL injection, since the engine compiles statements before parameters are bound to the database.  This means that, even in the case of users passing an otherwise valid SQL injection, the system will still treat it purely asn a string.

In Kwil, all bind parameters that accept user inputs must begin with a "$".  For global variables, such as a transaction caller's wallet address, bind parameters must begin with "@".

```go
type ExpressionBindParameter struct {
    Parameter string
}
```

_In SQL:_

```sql
UPDATE users
SET username = /*user bind parameter*/ $new_username,
age = /*user bind parameter*/ $new_age
WHERE wallet_address = /*global bind parameter*/ @caller
AND id = /*user bind parameter*/ $target_id;
```

_This statement allows a user to update a user record's username and age only if the value in the "wallet_address" column matches their address._

#### ExpressionColumn

A column expression is used to specify a column that should be used in a query.  It can also take a table name, which is necessary if you are using multiple tables (e.g. in a statement containing a join).

```go
type ExpressionColumn struct {
    Table  string
    Column string
}
```

_In SQL:_

```sql
SELECT *
FROM followers
INNER JOIN users
## Expressions containing table.column:
ON users.id = followers.followed_id
WHERE followers.follower_id = $user_id;
```

#### ExpressionBinaryComparison

A binary comparison expression is used to compare two expressions against each other.  This is commonly used in WHERE predicates to specify desired result attributes.

The expressions are compared using a binary operator (TODO: link to operators page).

```go
type ExpressionBinaryComparison struct {
    Left     Expression
    Operator BinaryOperator
    Right    Expression
}
```

_In SQL:_

```sql
SELECT *
FROM users
## Binary comparison in a where clause:
WHERE age > 20
```

#### ExpressionFunction

A function is used to execute some pre-defined functionality on a value or set of values.  A full list of Kwil's supported functions can be found here: TODO: link

```go
type ExpressionFunction struct {
    Function SQLFunction
    Inputs   []Expression
}
```

_In SQL:_

```sql
INSERT INTO users (username, age)
## Using an "abs" function to get the absolute value of an integer
VALUES ($new_username, abs($new_age));
```

#### ExpressionList

An expression list is simply a list of other expressions, wrapped in parenthesis and delimited by commas.

```go
type ExpressionList struct {
    Expressions []Expression
}
```

_In SQL:_

```sql
SELECT *
FROM users
## A list of expression literals, within a binary comparison
WHERE username IN ('satoshi', 'hal_finney', 'roger_ver');
```

#### ExpressionCollate

A collation is used to specify the bitset that should be used when comparing values.  Kwil has 3 predefined collation types, and currently does not support custom collations.  The 3 supported are:

- BINARY: Compares values based on their binary representation, which makes string comparisons case-sensitive and accent-sensitive.
- RTRIM: Compares strings while ignoring trailing whitespace.
- NOCASE: Compares strings case-insensitively.

```go
type ExpressionCollate struct {
    Expression Expression
    Collation  CollationType
}
```

_In SQL:_

```sql
SELECT *
FROM users
WHERE username = 'sAtOshI'
## case-insensitive collation:
COLLATE NOCASE;
```

#### ExpressionStringCompare

A string comparison expression is used to compare two strings against each other.  It also has an escape clause, which is used to escape characters.  The escape clause can only be used with LIKE and NOT LIKE string operators.

For a full list of string operators, see the operators page TODO: link to operators.

```go
type ExpressionStringCompare struct {
    Left     Expression
    Operator StringOperator
    Right    Expression
    Escape   Expression // can only be used with LIKE or NOT LIKE
}
```

_In SQL:_

```sql
SELECT *
FROM users
## LIKE comparison with ESCAPE clause
WHERE username LIKE '%\_finney' ESCAPE '\';
```

#### ExpressionIsNull

An IsNull expression is used to determine whether or not the result of some expression is null.

```go
type ExpressionIsNull struct {
    Expression Expression
    IsNull     bool
}
```

_In SQL:_

```sql
SELECT *
FROM posts
## getting all posts that have a body
WHERE content NOT NULL
```

#### ExpressionDistinct

A distinct expression is used to determine whether or not two expressions evaluate to distinct results from one another.

```go
type ExpressionDistinct struct {
    Left     Expression
    Right    Expression
    IsNot    bool
    Distinct bool
}
```

_In SQL:_

```sql
SELECT *
FROM users
# Getting all users distinctly different from satoshi.
WHERE username IS DISTINCT FROM 'satoshi';
```

#### ExpressionBetween

A between expression is used to filter values that are between the results of two other expressions.

```go
type ExpressionBetween struct {
    Expression Expression
    NotBetween bool
    Left       Expression
    Right      Expression
}
```

_In SQL:_

```sql
SELECT *
FROM users
# filtering results of a column between two integer literals
WHERE age BETWEEN 18 AND 30;
```

#### ExpressionIn

An In expression is used to specify values that are within some set of values.

```go
type ExpressionIn struct {
    Expression    Expression
    NotIn         bool
    InExpressions []Expression
}
```

_In SQL:_

```sql
SELECT *
FROM posts
# Getting posts from a set of authors
WHERE author_id IN ('satoshi', 'hal_finney', 'roger_ver');
```

#### ExpressionSelect

The select expression, also known as a subquery, allows users to specify results of a SELECT that should be used in a query.  In most cases, this has to be a "scalar subquery"; that is, it can only return one column.  This is not enforced at the statement validation level, but instead at the execution level.

```go
type ExpressionSelect struct {
    IsNot    bool
    IsExists bool
    Select   *SelectStmt
}
```

_In SQL:_

```sql
INSERT INTO posts (id, title, content, author_id)
# A scalar subquery used to get the user_id of the caller
VALUES ($id, $title, $content, (
    SELECT id
    FROM users
    WHERE wallet_address = @caller
    )
);
```

#### ExpressionCase

A case expression is the logical equivalent of an IF... THEN... ELSE... statement, where conditions are checked until one is met.

```go
type ExpressionCase struct {
    CaseExpression Expression
    WhenThenPairs  [][2]Expression
    ElseExpression Expression
}
```

_In SQL:_

```sql
SELECT username, 
# If users have an age, we will identify the age classification
    CASE age NOT NULL
        WHEN age > 65 THEN 'Senior'
        WHEN age > 35 THEN 'Middle-Age'
        WHEN age > 18 THEN 'Young-Adult'
        ELSE 'Minor'
    END
AS age_classification
FROM users;
```
