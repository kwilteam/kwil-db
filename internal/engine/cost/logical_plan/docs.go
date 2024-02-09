/*
Package logical_plan contains the logical plan of the query engine.

A LogicalPlan is a tree of operators that represents the logical steps to
execute a query.

The hierarchy of the logical operators is:
limit
  - project
  - sort
  - aggregate/distinct
  - aggregate/having
  - aggregate/group
  - filter/where
  - scan
*/
package logical_plan
