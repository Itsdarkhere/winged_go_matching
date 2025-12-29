package conn

const (
	sqlboilerTomlMySQL = `output = "%s"
pkgname = "%s"

[mysql]
host = "%s"
port = %d
user = "%s"
pass = "%s"
dbname = "%s"
sslmode = "skip-verify"
`
)
