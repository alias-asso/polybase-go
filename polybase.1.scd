POLYBASE(1) "git.sr.ht/~alias/polybase-go" "General Commands Manual"

# NAME

*polybase* - Manage polybase database from the command line

# SYNOPSIS

*polybase* [-db <PATH>] <command> [ARGUMENT]

# DESCRIPTION

*polybase* is a command-line interface for managing course data in the polybase
database. It provides commands for creating, reading, updating and deleting
course entries, as well as managing course quantities and visibility states.

# OPTIONS

- *-db* <PATH>  Path to database file (default: /var/lib/polybase/polybase.db)
- *-h*          Print help information
- *-v*          Print version information

# COMMAND

*create* <CODE> <KIND> <PART> [OPTIONS]
	Create a new course entry.

	Options:
	- *-n* <NAME>      Course name (required)
	- *-q* <QUANTITY>  Initial quantity (required)
	- *-t* <TOTAL>     Total quantity (default: same as quantity)
	- *-s* <SEMESTER>  Semester (required)
	- *-json*          Output in JSON format

*get* <CODE> <KIND> <PART>
	Display details for a specific course


*update* <CODE> <KIND> <PART> [OPTIONS]
	Update course information

	Options:
	- *-c* <CODE>      Update course code
	- *-k* <KEY>       Update course key
	- *-p* <PART>      Update course part
	- *-n* <NAME>      Update course name
	- *-q* <QUANTITY>  Update quantity
	- *-t* <TOTAL>     Update total quantity
	- *-s* <SEMESTER>  Update semester
	- *-json*          Output in JSON format

*delete* <CODE> <KIND> <PART>
	Remove a course from the database

*list* [OPTIONS]
	List all courses

	Options:
	- *-a*             Show hidden courses
	- *-s* <SEMESTER>  Filter by semester
	- *-c* <CODE>      Filter by code prefix
	- *-k* <KIND>      Filter by kind
	- *-p* <PART>      Filter by part number
	- *-json*          Output in JSON format

*quantity* <CODE> <KIND> <PART> <DELTA>
	Update course quantity by adding DELTA (can be negative)

	Options:
	- *-json*          Output in JSON format

*visibility* <CODE> <KIND> <PART> [-s STATE]
	Set course visibility

	Options:
	- *-s*             Set visibility state (default: true)
	- *-json*          Output in JSON format

*help* [COMMAND]
	Show help message for a specific command

# EXAMPLES

Create a new course:
```
$ polybase create LU3IN009 TD 1 -n "Systèmes de Gestion de Bases de Données" -q 60 -s "S1"
```

Get course details:
```
$ polybase get LU2IN018 Memento 1
```

Get course details in JSON format:
```
$ polybase get LU2IN018 Memento 1 -json
```

Update course properties:
```
$ polybase update LU2IN018 TME 1 -q 20 -t 230
```

Update course identity:
```
$ polybase update LU2IN005 TD 1 -c LU2IN006 -k TD -p 2
```

List all courses including hidden:
```
$ polybase list -a
```

List all courses in JSON format:
```
$ polybase list -json
```

Set course visibility:
```
$ polybase visibility LU2IN018 TME 1 -s false
```

Modify course quantity:
```
$ polybase quantity LU2IN018 TME 1 -6
```

Delete a course:
```
$ polybase delete MU4IN600 TD 2
```

# FILES

*/var/lib/polybase/polybase.db*
	Default database location

# BUGS

Bug reports and feature requests should be submitted to:
https://git.sr.ht/~alias/polybase-go

# AUTHORS

Written by ALIAS (2024).
Licensed under TODO.
