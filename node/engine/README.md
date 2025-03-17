# Kwil Engine

This document outlines Kwil's engine, which is responsible for handling all database-related functionality.

The engine has the following responsibilities:

- Accepting DDL (`CREATE TABLE`, `CREATE ACTION`, etc.) statements, converting them to a structured format which can be held in memory,
and persisting them within the DB.
- Accepting SQL statements, parsing them, rewriting them to be deterministic, and executing them against the database.
- Executing actions that have been defined with a `CREATE ACTION` statement.
- Storing rules and enforcing access control rules for them.
- Managing developer-defined precompiles (extensions).
- Making all of the above operations deterministic.

## How The Engine Works

The engine has two functionalities: `execute` and `call`. `execute` is used when executing a raw statement, such as a SQL statement or a DDL statement. `call` is used when executing an action that has been defined either using a `CREATE ACTION` statement or as a precompile. When both `execute` and `call` are used, they are run against the [`*baseInterpreter`](./interpreter/interpreter.go) struct, which holds important metadata in-memory for the lifetime of the node.

The interpreter "interprets" a statement passed to `execute` by parsing the statement and traversing the AST, executing logic specific to each node in the AST's tree as it is encountered. It does this by converting statements (or parts of statements) into one of three functions: `execFunc`, `stmtFunc`, `exprFunc`:

```go
// execFunc is a block of code that can be called with a set of ordered inputs.
// For example, built-in SQL functions like ABS() and FORMAT(), or user-defined
// actions, all require arguments to be passed in a specific order.
type execFunc func(exec *executionContext, args []value, returnFn resultFunc) error

// stmtFunc is a block of code that executes a "statement" from the AST.
// "statements" are language features such as:
// - sql: INSERT/UPDATE/DELETE/SELECT
// - ddl: CREATE/ALTER/DROP
// - action logic: FOR loops / IF clauses / variable assignment
type stmtFunc func(exec *executionContext, fn resultFunc) error

// exprFunc is a function that returns a single value.
// It is used to represent pieces of logic that should evaluate to
// exactly one thing (e.g. arithmetic, comparison, etc.)
type exprFunc func(exec *executionContext) (value, error)
```

Notice that a `resultFunc` is passed around to `execFunc` and `stmtFunc` functions. The `resultFunc` allows the interpreter
to progressively write results while the interpreter executes. This means that if an action or statement returns many rows
of data, the interpreter will only read each row (and perform subsequent execution logic) as needed. In previous versions of
Kwil which relied on the PL/pgSQL interpreter, it would read _all_ requisite data from disk before processing it. The `resultFunc`
allows us to avoid this.

### Understanding `execute`

`execute` takes a statement (or a group of statements delimited by `;`), converts them into `stmtFunc`s, and executes them.
The implementation of each `stmtFunc` depends on the statement passed, but there are generally 4 types of stmtFuncs.

- SQL: an INSERT/UPDATE/DELETE/SELECT statement that is immediately executed against the database. More information on SQL statements
is included below in this document.
- DDL: any DDL that does something like creating/alterting/deleting a table, index, role, etc. The implementation for
DDL statements are all different and very implementation-specific, but also quite simple.
- CREATE ACTION: while technically a DDL, creating an action involves several extra concepts not used in other DDL. It converts each statement within
the action body into a reusable `stmtFunc`, and then wraps them into a single `execFunc`, which is cached and can be reused later.
- USE <extension>: also technically a DDL, but performs special logic involving extra concepts to initialize a developer-defined extension. It wraps
the user-defined behavior into an `execFunc`, which is cached and can be reused later.

**Note:** these 4 types of stmtFuncs are not explicitly defined anywhere in the code. It is simply a broad characterization,
and there are some statements that do not fall into any 4 of these (e.g. `SET CURRENT NAMESPACE`).

### Understanding `call`

The interpreter's `call` functionality allows a user to execute either an action or an extension method. It does this by accessing the locally cached `execFunc` for the action or extension method.

### SQL Queries

There are special considerations that the engine takes into account when executing SQL queries. These considerations are applicable both for ad-hoc queries during `execute`, and for queries within an action / extension.

By default, SQL queries are _not_ deterministic. I won't list all forms of non-determinism here, but one basic example is the order of returned results; `SELECT * FROM table` can return rows in any order. Since this breaks the determinism requirements of Kwil, Kwil rewrites queries to be deterministic (e.g. guaranteeing ordering). It does this by converting the SQL statements to a "logical plan".

A "logical plan" is a mathematical representation of operations applied on a "relation" (a table). It is based off of relational algebra; it is highly recommended that you learn basic relational algebra if you need to understand Kwil's query planner. Kwil converts a SQL statement to our own modified version of relational algebra, identifies areas of non-determinism, applies extra logic on these areas to make them deterministic, and then re-generates Postgres-compliant SQL. If a query doesn't need to be deterministic (if it is being used outside of consensus by a read-only RPC call), it will not have this additional logic applied.

## Structure of the Engine Code

The engine code has the following structure:

- `/`: The root directory (which this README is contained in) contains common pieces of code used throughout the rest of the subdirectories in the engine. These are primarily common types, errors, and lists of constants (e.g. functions).
- `/interpreter`: The main entry-point for the engine is in the `interpreter` package. This defines the logic for how statements are interpreted, how
data is stored on disk and represented in-memory, and how other packages are used and called. If you are new to this section of the code, I would recommend starting in `/interpreter/interpreter.go`, and branching out to other files and packages that are used within there.
- `/parse`: The `parse` package implements the parser for all of SQL Smart Contracts. It uses [Antlr v4](<https://www.antlr.org/>) as a parser-generator, and defines the languages AST, grammar rules, and other basic syntax validations.
- `/pg_generate`: The `pg_generate` package is a very simple package that allows ghenerating Postgres-compatible SQL from Kwil's SQL AST.
- `/planner`: The `planner` package implements Kwil's deterministic query planner. It has two sub-packages: `logical` and `optimizer`. The `optimizer` package is currently unused. The `logical` package contains the logical query planner.
