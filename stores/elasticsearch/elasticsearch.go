package elasticsearch

import (
	"github.com/olivere/elastic/v6"
)

const (
	Timestamp    = "timestamp"
	ServiceName  = "service_name"
	ResponseCode = "response_code"
	ResponseTime = "response_time"
)

type ES struct {
	client         *elastic.Client
	indexName      *string
	maxNumServices *int
}

func New(client *elastic.Client, indexName *string, maxNumServices *int) *ES {
	return &ES{
		client:         client,
		indexName:      indexName,
		maxNumServices: maxNumServices}

}
