# Kuneiform Interpreter

The Kuneiform interpreter is meant to be a simple interpreter for performing basic arithmetic and access control logic. It is capable of:

- if/then/else statements
- for loops
- basic arithmetic
- executing functions / other actions
- executing SQL statements

For all function calls, it will make a call to Postgres, so that it can 100% match the functionality provided by Postgres. This is obviously very inefficient,
but it can be optimized later by mirroring Postgres functionality in Go. For now, we are prioritizing speed of development and breadth of supported functions.