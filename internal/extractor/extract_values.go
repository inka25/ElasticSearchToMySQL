package extractor

import (
	"encoding/json"
	"strings"

	"github.com/tidwall/gjson"
)

func ExtractValues(result []byte, startTime, statusCode *string) [][]interface{} {

	parseResult := gjson.GetBytes(result, "range_aggs.buckets.0.terms_aggs.buckets").Array()
	values := [][]interface{}{}

	for i := 0; i < len(parseResult); i++ {

		serviceName := parseResult[i].Get("key").String()
		filterAggs := parseResult[i].Get("filter_aggs")
		count := filterAggs.Get("stats_aggs.count").Int()

		var statusString strings.Builder
		statusString.WriteString(*statusCode)
		if *statusCode == "2" || *statusCode == "3" || *statusCode == "4" {
			statusString.WriteString("xx")
		}

		var sum interface{}
		var avg interface{}
		sum = nil
		avg = nil
		if count != 0 {
			sum = filterAggs.Get("stats_aggs.sum").Num
			avg = filterAggs.Get("stats_aggs.avg").Num
		}

		percentiles := filterAggs.Get("percentiles_aggs.values").Value()
		bytePercent, _ := json.Marshal(percentiles)

		value := []interface{}{
			serviceName,
			*startTime,
			statusString.String(),
			count,
			sum,
			avg,
			bytePercent}

		values = append(values, value)
	}

	return values
}
