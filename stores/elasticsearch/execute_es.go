package elasticsearch

import (
	"context"

	"github.com/olivere/elastic/v6"
)

func (es *ES) ExecuteES(start, end, statusCode *string) (*elastic.SearchResult, error) {

	searchService := es.client.Search(*es.indexName).Size(0)
	aggs := esQueryBuilder(start, end, statusCode, es.maxNumServices)
	searchService = searchService.Aggregation("range_aggs", aggs)

	result, err := searchService.Do(context.Background())
	if err != nil {
		return nil, err
	}

	return result, nil
}

func esQueryBuilder(startTime, endTime, statusCode *string, maxNum *int) elastic.Aggregation {

	rangeAggs := elastic.NewRangeAggregation().
		Field(Timestamp).
		AddRange(*startTime, *endTime).
		SubAggregation("terms_aggs", elastic.NewTermsAggregation().
			Field(ServiceName).
			Size(*maxNum).
			SubAggregation("filter_aggs", elastic.NewFilterAggregation().
				Filter(elastic.NewPrefixQuery(ResponseCode, *statusCode)).
				SubAggregation("stats_aggs", elastic.NewStatsAggregation().
					Field(ResponseTime)).
				SubAggregation("percentiles_aggs", elastic.NewPercentilesAggregation().
					Field(ResponseTime)),
			),
		)

	return rangeAggs

}
