package extractor

import (
	"log"
	"testing"
)

const (
	ESSuccessResult1 = `
	{
		"range_aggs": {
			"buckets": [
				{
					"key": "2020-04-07 00:00:00.000-2020-04-07 23:59:59.999",
					"from": 1.5862176E12,
					"from_as_string": "2020-04-07 00:00:00.000",
					"to": 1.586303999999E12,
					"to_as_string": "2020-04-07 23:59:59.999",
					"doc_count": 66773278,
					"terms_aggs": {
						"doc_count_error_upper_bound": 2243429,
						"sum_other_doc_count": 57564899,
						"buckets": [
							{
								"key": "quest-be-service",
								"doc_count": 9208379,
								"filter_aggs": {
									"doc_count": 9204395,
									"percentiles_aggs": {
										"values": {
											"1.0": 0.0,
											"5.0": 0.0,
											"25.0": 2.0,
											"50.0": 13.688597258676,
											"75.0": 43.99206276685738,
											"95.0": 103.0,
											"99.0": 159.69074048378073
										}
									},
									"stats_aggs": {
										"count": 9204395,
										"min": 0.0,
										"max": 5240.0,
										"avg": 33.09449105563158,
										"sum": 3.04614768E8
									}
								}
							}
						]
					}
				}
			]
		}
	
	}
	`
	ESSuccessResult2 = `
	{
		"range_aggs": {
		"buckets": [
			{
				"key": "2020-04-07 00:00:00.000-2020-04-07 23:59:59.999",
				"from": 1.5862176E12,
				"from_as_string": "2020-04-07 00:00:00.000",
				"to": 1.586303999999E12,
				"to_as_string": "2020-04-07 23:59:59.999",
				"doc_count": 66773278,
				"terms_aggs": {
					"doc_count_error_upper_bound": 1233704,
					"sum_other_doc_count": 48604590,
					"buckets": [
						{
							"key": "quest-be-service",
							"doc_count": 9208379,
							"filter_aggs": {
								"doc_count": 9204395,
								"percentiles_aggs": {
									"values": {
										"1.0": 0.0,
										"5.0": 0.0,
										"25.0": 2.0,
										"50.0": 13.971458746105036,
										"75.0": 43.990980480827524,
										"95.0": 103.0,
										"99.0": 159.70118510391956
									}
								},
								"stats_aggs": {
									"count": 9204395,
									"min": 0.0,
									"max": 5240.0,
									"avg": 33.09449105563158,
									"sum": 3.04614768E8
								}
							}
						},
						{
							"key": "buyer-mission-service",
							"doc_count": 8960309,
							"filter_aggs": {
								"doc_count": 8899766,
								"percentiles_aggs": {
									"values": {
										"1.0": 0.0,
										"5.0": 0.0,
										"25.0": 1.0,
										"50.0": 4.0,
										"75.0": 38.363277289115686,
										"95.0": 206.6377380721109,
										"99.0": 1077.2108035471244
									}
								},
								"stats_aggs": {
									"count": 8899766,
									"min": 0.0,
									"max": 10003.0,
									"avg": 63.620279679263476,
									"sum": 5.66205602E8
								}
							}
						}
					]
				}
			}
		]
	}`
)

func TestExtractValues(t *testing.T) {

	var values = []struct {
		expression string
		length     int
	}{
		{ESSuccessResult1, 1},
		{ESSuccessResult2, 2},
	}

	startTime := "2006-02-01"
	statusCode := "1"

	for _, tt := range values {
		t.Run(tt.expression, func(t *testing.T) {
			actual := ExtractValues([]byte(tt.expression), &startTime, &statusCode)

			log.Println(len(actual), tt.length)
			if len(actual) != tt.length {
				t.Errorf("got %d want %d", len(actual), tt.length)
			}
		})
	}
}
