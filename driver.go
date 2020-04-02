package migrations

type Driver interface {
	CreateVersionsTable() error
	HasExecuted(version string) bool
	Up(migration Migration) error
	Down(migration Migration) error
}

