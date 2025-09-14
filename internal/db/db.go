package db

type DB struct {
	dsn string
}

func Open(dsn string) (*DB, error) {
	return &DB{dsn: dsn}, nil
}

func AutoMigrate(db *DB) error {
	return nil
}
