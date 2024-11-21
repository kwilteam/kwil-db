# internal

The packages in this top level internal are only for use by the main module
packages and apps like kwild.

These will be part of the kwil-db main module. As such other kwil-db module
cannot use them.

Any reusable code that we want to make publicly accessible may be moved out of
internal. Doing so means it should be documented for public consumption and
supported, while minimizing breaking changes.
