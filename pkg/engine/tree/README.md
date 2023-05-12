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

```sql
SELECT *
FROM users
WHERE username = 'sAtoSHi'
## Start of collation
COLLATE NOCASE;
## End of collation
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
## Start of conflict target
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
## Start of CTE
WITH users_followers AS (
    SELECT followed_id AS followers
    FROM followers
    WHERE follower_id = (
        SELECT id FROM users
        WHERE username = 'satoshi'
    )
)
## End of CTE
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
## Start of DeleteStmt
DELETE FROM users
WHERE username = 'sbf'
RETURNING (
    SELECT *
    FROM users
    LIMIT 10
);
## End of DeleteStmt
```

#### Expression

Expressions are combinations of one or more values, identifiers, comparisons, and other expressions that evaluate to a value / set of values.  There are multiple types of expressions in Kwil, so they have [their own page](./expression.md).