package msql

import "github.com/jmoiron/sqlx"

type MSQL struct {
	client    *sqlx.DB
	tableName *string
}

func New(client *sqlx.DB, tableName *string) *MSQL {
	return &MSQL{
		client:    client,
		tableName: tableName,
	}
}
