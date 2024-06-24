/*
Package cost provides a cost model for SQL.

The model this package uses is based on https://howqueryengineswork.com, I
recommend reading this book before reading this package.

Some packages in this directory should be moved out to maybe the engine package,
like `catalog` and `datasource`. They are here for now because current implementation
mocked the data source and catalog, and hasn't implemented the real data source yet.

The `catalog` package defines the interface for query planner(logical plan) to
get correspond DataSource of a table. This package is here only

The `costmodel` package was what I think where to apply the cost calculation logic.
It will build an optimized plan tree and track the statistics(or transform) while
building the plan tree. It will get a cost after build the plan tree. This is
where I left off.

The `datasource` package provides the data source for query planner. It defines
the interface to get schema/scan/statistics information from the underlying data
source. It also implements two types of data sources for testing purposes:
`memDataSource` and `CsvDataSource`.

The `datatypes` package defines the data types used in logical plan and cost model.

The `demo` pakcage is a demo for how a SQL got planned and executed. It helped me
to understand the whole lifetime of a SQL query. It uses virtual plan to actual
get the data out from somewhere(like a CSV file).

The `internal` package has some helper functions for testing.

The `logical_plan` package defines the logical plan, it has it's own expressions
to represent the transformation of the columns. It also defines several operators
to represent the logical plan.

The `memo` package is a fail attempt to use the memoization technique used in most of
the modern database query optimizers, simplified just for this project. TBH I'm
still not clear how to implement it, so I decided to leave it and do the cost estimation
without it.

The `optimizer` package is where I planned to transform the logical plan to a physical
(virtual) plan. It also implements(partial) two very simple yet very general optimization
rules to optimize the logical plan, which are common in all modern database query
optimizers. The two rules are predicate pushdown and projection pushdown.

The `plantree` package is a tree structure to represent a plan tree. It has functions
to traverse the tree and do transformations on the tree as well.  I planned to
implement the visitor pattern on this structure, but ended up just using recursive
'builder' functions(in `query_planner` package) to build the plan tree. I still
think the visitor pattern will become handy in the future.

The `query_planner` package is the main package to build the plan tree. It transforms
a kuneiform AST to a logical plan(another tree). With our new AST and Analyzer,
some code in this pkg could be implemented in the Analyzer.

The `virtual_plan` package defines the virtual plan, which I originally plan to use
to do the cost estimation on. It's now only for demo purpose.

All tests will use test data from the testdata directory.
*/
package cost
