package app

import "codeberg.org/mahlzeit/mahlzeit/db/queries"

// Application holds the required resources for the application.
// This includes, but is not limited, to loggers, databases, cache handlers, etc.
type Application struct {
	Templates Templates
	Queries   *queries.Queries
}

type Configuration struct {
	Database struct {
		ConnectionString string `toml:"connection-string"`
	} `toml:"database"`
	Web struct {
		Endpoint    string `toml:"endpoint"`
		TemplateDir string `toml:"template-dir"`
	}
}
