package migrations

import (
	"embed"
	"io/fs"
)

//go:embed sql/*.sql
var embedded embed.FS

func FS() (fs.FS, error) {
	return fs.Sub(embedded, "sql")
}
