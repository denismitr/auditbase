package db

type Migrator interface {
	Up() error
	Down() error	
}