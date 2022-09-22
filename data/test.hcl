schema "example" {
    name = "example"
}

table "users" {
  schema = schema.example
  column "id" {
    nullable = false
    type = int
  }
  column "name" {
    nullable = true
    type = string(100)
  }
  primary_key {
    columns = [column.id]
  }
}

table "blog_posts" {
  schema = schema.example

  column "id" {
    nullable = false
    type = int
  }

  column "title" {
    nullable = true
    type = string(100)
  }

  column "body" {
    nullable = true
    type = string
  }

  column "author_id" {
    nullable = true
    type = int
  }

  primary_key {
    columns = [column.id]
  }

  foreign_key "author_fk" {
    columns     = [column.author_id]
    ref_columns = [table.users.column.id]
  }

  index "author_id" {
    unique  = false
    columns = [column.author_id]
  }
}