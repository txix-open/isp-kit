package dbx

type Option func(db *Client)

func WithMigration(migrationDir string) Option {
	return func(db *Client) {
		db.withMigration = true
		db.migrationDir = migrationDir
	}
}
