// Package db embeds the SQL migrations so tests and tools can apply them
// without depending on the working directory.
package db

import "embed"

//go:embed migrations/*.sql
var Migrations embed.FS

// LocalRoles creates the database roles for local development and tests.
//
//go:embed local/roles.sql
var LocalRoles string
