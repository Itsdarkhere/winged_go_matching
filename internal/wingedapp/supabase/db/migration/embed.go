package migration

import "embed"

//go:embed *.sql
var Stmts embed.FS
