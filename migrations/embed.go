package migrationsfs

import "embed"

//go:embed *.sql
var PostgresMigrations embed.FS
