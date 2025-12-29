package conn

const (
	ReadyForDeletion = "ready-for-deletion"

	sqlStmtSelectPidsWhereAppName = `
		SELECT pid, application_name FROM pg_stat_activity
		WHERE application_name = $1
	`

	sqlStmtCountPids = `
		SELECT COUNT(*) as count FROM pg_stat_activity
	`
)
