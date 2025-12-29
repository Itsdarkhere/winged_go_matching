package conn

type PGDB struct {
	Name string `json:"name" db:"name"`
}

type PGDBs []PGDB

func (p PGDBs) StrArr() []string {
	s := make([]string, 0, len(p))
	for _, db := range p {
		s = append(s, db.Name)
	}
	return s
}

type PID struct {
	Pid             int    `db:"pid"`
	ApplicationName string `db:"application_name"`
}
