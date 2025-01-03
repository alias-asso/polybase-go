POLYBASED(1) "git.sr.ht/~alias/polybase-go" "General Commands Manual"

# NAME

*polybased* - Manage polybase database from the web browser

# SYNOPSIS

*polybased* [OPTIONS]

# DESCRIPTION

*polybased* is a web server application that provides an interface for managing
education courses and their resources. It features both public and
administrative views, LDAP-based authentication, and real-time course inventory
managment. The system is designed to help track course materials, manage
visibility, and handle course quantities through a responsive web interface.

# OPTIONS

- *-c* <path>  Path to config file (default: /etc/polybase/config.cfg)
- *-v*         Print version information
- *-h*         Print this help message

# CONFIGURATION

*polybased* uses a configuration file, its default location for this file is
/etc/polybase/polybase.cfg. The configuration file uses TOML format and supports
the following sections and parameters:

## server

*host*
	Host on which the server will be hosted (default: 0.0.0.0)

*port*
	Port on which the server will be hosted (default: 1265)

*static*
	Location where the static assets are stored (default: /var/www/polybase/static)

## database

*path*
	Path to the polybase database (default: /var/lib/polybase/polybase.db)

## ldap

*host*
	Host of the LDAP server (default: 127.0.0.1)

*port*
	Port of the LDAP server (default: 389)

*user_dn*
	User domain name of the LDAP server (must contain %s for username)

## auth

*jwt_secret*
	Secret for JWT authorization

*jwt_expiry*
	Duration before JWT expiry (default: "4320h")

# DATABASE SCHEMA

The application uses SQLite with the following main table structure:

## Course Table

*code*
	Course code identifier (TEXT, PRIMARY KEY part 1)

*kind*
	Course type or category (TEXT, PRIMARY KEY part 2)

*part*
	Section number (INTEGER, PRIMARY KEY part 3)

*parts*
	Total number of sections (INTEGER)

*name*
	Course name (TEXT)

*quantity*
	Current available quantity (INTEGER)

*total*
	Total capacity (INTEGER)

*shown*
	Visibility flag (INTEGER, 0 or 1)

*semester*
	Academic semester (TEXT)

# WEB ENDPOINTS

## Public Endpoints

*GET /*
	Public view of visible courses

*GET /login*
	Login page

*POST /auth*
	Authentication endpoint

## Protected Endpoints

*GET /admin*
	Administrative dashboard

*GET /admin/courses/new*
	New course creation form

*GET /admin/courses/edit/{code}/{kind}/{part}*
	Course editing form

*GET /admin/courses/delete/{code}/{kind}/{part}*
	Course deletion form

*PUT /admin/courses/{code}/{kind}/{part}*
	Update course information

*DELETE /admin/courses/{code}/{kind}/{part}*
	Delete a course

*PATCH /admin/courses/{code}/{kind}/{part}/quantity*
	Update course quantity

*PATCH /admin/courses/{code}/{kind}/{part}/visibility*
	Toggle course visibility

# AUTHENTICATION

The system uses LDAP for authentication and JWT tokens for session management.
The authentication flow is as follows:

1. Use submits credentials via the login form
2. Server validates credentials against LDAP server
3. On success, a JWT is generated and stored as an HTTP-only cookie
4. Token is validated for all protected access

# FILES

*/etc/polybase/polybase.cfg*
	Default configuration file location

*/var/lib/polybase/polybase.db*
	Default database location

*/var/www/polybase/static*
	Default static files location

# BUGS

Bug reports and feature requests should be submitted to:
https://git.sr.ht/~alias/polybase-go

# AUTHORS

Written by ALIAS (2024).
Licensed under TODO.
