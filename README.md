# Pldoc -  documentation generator for PL/SQL

**Note**: pldoc is under development.

Pldoc is a documentation generator for PL/SQL that is inspired by how documentation
is generated in the Go programming language. It takes source code of specification files for
PL/SQL packages and generates documentation in html format. Generated documentation can
be placed on a server or viewed locally. The source of documentation is the comments in
code. The comment that hasn't empty lines between itself and a code object (function, variable etc)
is considered as the object's documentation. For example, the next code snippet is
the function `has_permission` with its documentation:

```
-- Returns true if the user has access to perform
-- this operation.
function has_permission(user_id number, op varchar2) return boolean;
```

If there is a blank line (or lines) between a comment and an object, the comment
isn't processed as documentation:

```
-- Returns true if the user has access to perform
-- this operation.

function has_permission(user_id number, op varchar2) return boolean;
```

However, an empty line (or lines) that is the part of a comment, isn't taken into account
as a separator between documentation and an object and a whole comment
is parsed as documentation:

```
-- Returns true if the user has access to perform
-- this operation.
--
function has_permission(user_id number, op varchar2) return boolean;
```

For now, pldoc can generate docs for:

- Package documentation
- Functions, procedures
- Cursors
- Records declaration
- Varrays, tables
- Constants and variables
 
## Limitations

- Specification files should be in UTF-8 encoding

## Build from source
```
go build pldoc.go
```

## Usage

This command generates documentation for each file with `.pks` extension
in the `source_directory`. The documentation files will be placed in the
`documentation` directory. 
```
pldoc --output=documentation source_directory
```

To change the extension of searched files, `ext` flag is used:

```
pldoc --output=documentation --ext=sql source_directory
```

Pldoc can generate docs from multiple directories, all you need is just list them with a space. Each
directory will be walked recursively, for example:

```
pldoc --output=documentation source_dir1 source_dir2 source_dir3
```
## Comment styles

It's better not to decorate you comments. Bad example:

```
/*****************************************
* This function returns the my_table row
* by its id
*****************************************/
```

Good example:

```
-- This function returns the my_table row
-- by its id
```

Another one good example:

```
/*
   This function returns the my_table row
   by its id
*/
```

## Preformatted text and lists

Any line that has greater indentation that comment's first line
is preformatted text.

```
-- This procedure returns formatted output of
-- passed record. It looks like this:
--     key1: value 1
--     key2: value 2
--     key3: value 3
```

In documentation, lines from 3 to 5 will be enclosed in
`pre` tag.
