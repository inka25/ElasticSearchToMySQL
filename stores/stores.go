package stores

import (
	"github.com/olivere/elastic/v6"
)

type ElasticSearch interface {
	ExecuteES(start, end, statusCode *string) (*elastic.SearchResult, error)
}

type MySQL interface {
	InsertToDB(values ...[]interface{}) error
}
