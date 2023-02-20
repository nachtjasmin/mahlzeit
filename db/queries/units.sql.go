// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.0
// source: units.sql

package queries

import (
	"context"
)

const getAllUnits = `-- name: GetAllUnits :many
select id, name
from units
order by name
`

func (q *Queries) GetAllUnits(ctx context.Context) ([]Unit, error) {
	rows, err := q.db.Query(ctx, getAllUnits)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Unit
	for rows.Next() {
		var i Unit
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}