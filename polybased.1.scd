POLYBASED(1) "github.com/alias-asso/polybase-go" "General Commands Manual"

# NAME

*polybased* - Manage polybase database from the web browser

# SYNOPSIS

*polybased* [OPTIONS]

# DESCRIPTION

*polybased* is a web server application that provides an interface for managing
education courses and their resources. It features both public and
administrative views, OIDC-based authentication, and real-time course inventory
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

## oidc

*client_id*
	OIDC client identifier

*client_secret*
	OIDC client secret

*issuer_url*
	OIDC issuer URL

*redirect_uri*
	OIDC callback URL registered with the provider

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
	Redirect to the configured OIDC provider

*GET /auth/callback*
	OIDC callback endpoint

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

The system uses OIDC for authentication and JWT tokens for session management.
The authentication flow is as follows:

1. User follows the login redirect
2. Server sends the browser to the configured OIDC provider
3. Provider returns to the callback URL with an authorization code
4. Server exchanges the code, verifies the ID token, and stores a JWT as an HTTP-only cookie
5. Token is validated for all protected access

# FILES

*/etc/polybase/polybase.cfg*
	Default configuration file location

*/var/lib/polybase/polybase.db*
	Default database location

*/var/www/polybase/static*
	Default static files location

# BUGS

Bug reports and feature requests should be submitted to:
https://github.com/alias-asso/polybase-go

# AUTHORS

Written by ALIAS (2024).
Licensed under TODO.
