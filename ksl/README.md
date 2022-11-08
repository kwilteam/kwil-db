# KSL Syntax Specification

This is the specification of the syntax and semantics of the native syntax for KSL. KSL is a system for defining schemas for applications.

## Structural Elements

The structural language consists of syntax representing the following constructs:

- _Directives_, which appear in the top-level part of a file and direct the parser.
- _Blocks_, which provide a grouping of elements by a type and optional modifiers and labels.
- _Block Content_, which consists of a collection of properties, declarations, child blocks, and annotations.
  - _Properties_, which are key-value pairs
  - _Declarations_, which declare a property name, type, optional size, and annotations
  - _Block Annotations_, which assign metadata to the enclosing block
  - _Child Blocks_, which are just blocks that appear in the body of another block

### Directives

A _directive_ is a top-level element that gives further instructions or clarifies context of a document.
```
@target "postgres"
@option include_files = true
@import "other.kwil"
```

### Blocks

A block is a logical grouping of elements that is annotated with a type, optional modifiers, and zero or more labels. Blocks create a structural hierarchy which can be interpreted by the calling application. A block label is a key-value pair. If the value is omitted, it's interpreted to be __true__ or __present__.

```
role general [default] {}
role admin extends general {}
```

In the above example, __role__ is the block type, __general__ is the block name, and __[default]__ is the label. There is also another role defined that __extends__ __general__.

### Declarations

A _declaration_ is a type of property that allows you to _define_ a field. It consists of a name followed by a colon, then a type specification, and zero or more _field annotations_.

A type specification consists of an optional array indicator, a name, and optional size, and an optional nullable specifier(?). All fields are considered not nullable unless the nullable specifier is included.

A _field annotation_ is an '@' sign followed by a label, with optional arguments.
```
table users {
    id:         int                   // an integer field
    age:        int(8)   @default(21) // an 8-bit integer field with a default value of 21
    shoe_sizes: []int                 // an integer array
    name:       string?               // an optional string field
}
```

### Properties
A _property_ a key-value pair in the body of a block.

```
datasource db {
    driver      = "postgres"
    schema_file = "kwil.schema"
}
```

### Literal Values

A _literal value_ immediately represents a particular value of a primitive type.

- Integer literals represent values of type _int64_.
- Float literals represent values of type _float64_.
- The `true` and `false` keywords represent values of type _bool_.
- The `null` keyword represents a null value of the dynamic pseudo-type.

### Collection Values

A _collection value_ combines zero or more other values to produce a collection value. Only list and object values can be directly constructed via native syntax.

- `{foo = "baz"}` is interpreted as an object with an attribute named _foo_ with a value of _"baz"_.
- `["hello", "world"]` is a list of strings

Between the open and closing delimiters of these sequences, newline sequences are ignored as whitespace.


### Block Annotations
A _block annotation_ is a special kind of _annotation_ that appears in the body of a block. It is denoted by two '@@' to differentiate it from a _field annotation_. A block annotation has a name and an optional argument list.

```
table users {
    id: int @pk
    name: string?

    @@index([name], ondelete=SET_NULL)
}
```

When defining an _annotation_ or _block annotation_ as a function, keyword arguments must follow any positional arguments. Positional arguments are not allowed after keyword arguments.