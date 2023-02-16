package app

import "codeberg.org/mahlzeit/mahlzeit/db/queries"

// Application holds the required resources for the application.
// This includes, but is not limited, to loggers, databases, cache handlers, etc.
type Application struct {
	Templates Templates
	Queries   *queries.Queries
}
