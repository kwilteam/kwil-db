# Transaction cost and pricing

## issue items

-[] update query_planner for insert/update/delete. Does howqueryengineswork.com describe how, or we just do it based on expected AST for these? buildInsertStmt, buildUpdateStmt
-[] complete `costmodel` functions for computing cost from `LogicalPlan`, similar to virtual_plan pkg
-[] implement a Catalog (GetDataSourcer for qp.NewPlanner)
-[] collect and maintain statistics
  -[] changesets at end can update ground truth stats
  -[] between statements (either within action or in different chain txns), stats change require audit table
-[] translate stats solutions inside of procedures

## plan tree tldr

a logical plan has: (1) expression and (2) input / source plan

e.g. SELECT * FROM employee WHERE state = 'CO'

Projection: #id, #first_name, #last_name, #state, #salary
    Filter: #state = 'CO'
        Scan: employee; projection=None

- Projection input is the filter plan, expression is the col names
- Filter input is the scan plan, expression is the EQ binary op
- Scan is at the bottom, has data source for a table

## types overview

datatypes.Statistics instance pertains to a *table*: RowCount and []ColumnStatistics.

DataSource is also for a table:

- Statistics() returns datatypes.Statistics
- Schema() returns datatypes.Schema

Catalog returns a DataSource for a table (TableDef).

In query_planner:

- query_planner.NewPlanner input is Catalog
- query_planner uses GetDataSource in buildTableSource
- Scan plan (a *ScanOp) created from the source, when buildSelectCore->buildFrom->buildRelation gets a tree.RelationTable

optimizer:

- pushdown rules: projection and predicate
- can still use these rules to transform a LogicalPlan
- Optimize() method used to make a virtual plan from a logical plan
- the original idea was to do optimization with a dp/memo search
- optimizer.NewPlanner was to convert logical plan tree to vp tree (and expressions)

costmodel:

- RelExpr / BuildRelExp / EstimateCost(*RelExpr)
- compute cost from LogicalPlan
- replaces initial approach in optimizer / virtual_plan

## statistics

- row counts
- MCVs
- min / max / range vals
  
some can maintain without full rescans:

- mean
- number of unique vals

## maintaining table / column value statistics

exact vs approximate

occasional full scan

### within actions

before executing each instruction, get cost estimate that uses plan and table stats

the actual execution should update statistics

- use changesets in replication stream at end commit to reestablish ground truth stats
- automatically using triggers and our own stats tables for live updates?  listen/notify with triggers? audit table with triggers
- extension to capture deltas?  https://github.com/mreithub/pg_recall

https://blog.sequin.io/all-the-ways-to-capture-changes-in-postgres/
https://github.com/dennwc/go-fdw
https://github.com/turbot/steampipe-postgres-fdw
https://github.com/Percona-Lab/clickhousedb_fdw

### within procedures

probably will create an extension that executes the procedure as represented by our ast.

- pass the procedure body from kwild to pg as... string argument to SQL extension call? Or just read by pg from some bytea column like schemas_content on dataset load?
- does the extension create the AST from pl body string with our parser code? or does kwild provide it to the extension (e.g. gob serializn)?
- something like the following? `SELECT * FROM kwil_pl_call(procIDorContent, args, costLimit);`

procedure needs access to a gas limit / allotment

before each statement (of a certain kind, SQL, functions), get cost estimate

- cost of stmt: talk to kwild, or use same cost code compiled in extension?
- where to get statistics? talk to kwild? have kwild write to ephemeral stats table before proc call?
- how to update statistics? triggers to write to audit table

use extension functions to run our code or talk to kwild?
(for stats or cost computation)
