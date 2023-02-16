package db

// We are using [sqlc] for the generation of the query code for the database.
// Since we are using PostgreSQL as the underlying database, it's possible to make use of all
// the fancy Postgres features by using the pgx/v4 driver. Although pgx/v5 would be theoretically
// possible, it's still experimental. Once [#1823] is fixed, we can switch to pgx/v5.
//
// [sqlc]: https://sqlc.dev
// [#1823]: https://github.com/kyleconroy/sqlc/issues/1823
//go:generate sqlc generate
