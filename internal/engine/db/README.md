# Metadata Store

The metadata store is responsible for storing, retrieving, and upgrading database metadata.

This document tracks the types of metadata and their changes (reflected in migrations)

## Types

- Table: a table in a database
- Procedure: a predefined set of statements to be executed (formerly known as "action")
- Extension: an imported / used extension within a database

### Table

Current Version: 1

### Procedure

Current Version: 2

#### Changelog

- Version 2: Changed the meaning of `public/private`.  `public=true` still means the same thing, however `public=false` now means that a [procedure] is private and cannot be called externally.  Previously, `public=false` meant that a procedure could only be called by the database owner.  This has been replaced with an `owner` modifier.

### Extension

Current Version 1