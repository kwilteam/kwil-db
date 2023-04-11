# Validation
Database schemas must adhere to a set of validation constraints.  Due to the vast number of possible database configurations,
all validation cases are outlined below.  Each rule is numbered and can be referenced in the code.  The ranges for rules and their sections
are as follows:

| Section | Range |
|---------|-------|
| General | 0-99 |
| Tables | 100-199 |
| Table | 200-299 |
| Columns | 300-399 |
| Column | 400-499 |
| Attributes | 500-599 |
| Attribute | 600-699 |
| SQL Queries | 700-799 |
| SQL Query | 800-899 |
| Parameter/Where Clauses | 900-999 |
| Parameter/Where Clause | 1000-1099 |
| Indexes | 1100-1199 |
| Index | 1200-1299 |
| Roles | 1300-1399 |
| Role | 1400-1499 |

The reason I have broken out rules into sections delineated by single vs multiple types (i.e. tables and table) is to identify which
validations need context of other members of the same type.  The file names are prepended with the section number to make it easier
to find the validation code for a specific rule.  While this is not exactly idiomatic, it makes it much easier to maintain the code,
and ensure that we cover all possible cases.

## General
| Rule | Description |
|------|-------------|
| 0 | Database name must be valid |
| 1 | Owner address must be valid |

## Tables
| Rule | Description |
|------|-------------|
| 100 | Table names must be unique |
| 101 | Must have at least 1 table |
| 102 | Must have less than max allowed tables |

## Table
| Rule | Description |
|------|-------------|
| 200 | Table name must be valid |
| 201 | Table name must not be a reserved word |

## Columns
| Rule | Description |
|------|-------------|
| 300 | Column names must be unique within the table |
| 301 | Column count must not exceed maximum |

## Column
| Rule | Description |
|------|-------------|
| 400 | Column name must be valid |
| 401 | Column name must not be a reserved word |
| 402 | Column type must be valid |

## Attributes
| Rule | Description |
|------|-------------|
| 500 | Attribute types must be unique in the column |
| 501 | Attribute count must not exceed maximum |
| 502 | Cannot have unique and default attribute on same column |

## Attribute
| Rule | Description |
|------|-------------|
| 600 | Attribute type must be valid |
| 601 | Attribute value must valid for attribute type |
| 602 | Default attribute must be valid for column type |
| 603 | Attribute must be applicable to column type |

## SQL Queries
| Rule | Description |
|------|-------------|
| 700 | SQL query names must be unique |
| 701 | SQL query count must not exceed maximum |

## SQL Query
| Rule | Description |
|------|-------------|
| 800 | SQL query name must be valid |
| 801 | SQL query type must be valid |
| 802 | Table must exist |
| 803 | Insert and update queries must have at least 1 parameter |
| 804 | Update and delete queries must have at least 1 where clause |
| 805 | Insert can not have a where clause |
| 806 | Delete can not have a parameter |
| 807 | All not-null columns in the table must be included in an insert |
| 808 | Name must not be reserved key-word |

## Parameter/Where Clauses
| Rule | Description |
|------|-------------|
| 900 | Parameter names must be unique within the query (for both parameters and where clauses) |
| 901 | A column can only be used in one parameter per SQL query |
| 902 | Parameter count must not exceed maximum |
| 903 | Where clause count must not exceed maximum |

## Parameter/Where Clause
| Rule | Description |
|------|-------------|
| 1000 | Parameter name must be valid |
| 1001 | Column must exist |
| 1002 | If not static, then value must be empty |
| 1003 | If modifier is Caller, then value must be empty |
| 1004 | If modifier is Caller, then param must be static |
| 1005 | If modifier is Caller, then column must be a string |
| 1006 | If modifier is Caller, then column must have not have min length > 42 |
| 1007 | If modifier is Caller, then column must have not have max length < 44 |
| 1008 | Operator must be valid |
| 1009 | Operator must be useable on column type |
| 1010 | Modifier value must be valid |

## Indexes
| Rule | Description |
|------|-------------|
| 1100 | Index names must be unique |
| 1101 | Index count must not exceed maximum |

## Index
| Rule | Description |
|------|-------------|
| 1200 | Index name must be valid |
| 1201 | Index name must not be a reserved word |
| 1202 | Index type must be valid |
| 1203 | Index table must exist |
| 1204 | Index column(s) must exist |
| 1205 | Index column(s) must be unique |
| 1206 | Index must have at least 1 column |
| 1207 | Index must have at most 3 columns |

## Roles
| Rule | Description |
|------|-------------|
| 1300 | Role names must be unique |
| 1301 | Role count must not exceed maximum |

## Role
| Rule | Description |
|------|-------------|
| 1400 | Role name must be valid |
| 1401 | Role permissions must exist as queries |
| 1402 | Role permissions must be unique within the role |