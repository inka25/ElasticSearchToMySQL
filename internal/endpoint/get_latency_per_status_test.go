package endpoint

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"service-log/mocks/storesmock"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/olivere/elastic/v6"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/tidwall/gjson"
)

const (
	ESSuccessResult1 = `
	{
		"took": 153334,
		"timed_out": false,
		"_shards": {
			"total": 4,
			"successful": 4,
			"skipped": 0,
			"failed": 0
		},
		"hits": {
			"total": 324327126,
			"max_score": 0.0,
			"hits": []
		},
		"aggregations": {
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
			}
		}
	}`
	ESWrongTimeFormat = `
	{
		"error": {
			"root_cause": [
				{
					"type": "parse_exception",
					"reason": "failed to parse date field [2020-04-07H00:00:00.000] with format [yyyy-MM-dd HH:mm:ss.SSS]"
				}
			],
			"type": "search_phase_execution_exception",
			"reason": "all shards failed",
			"phase": "query",
			"grouped": true,
			"failed_shards": [
				{
					"shard": 0,
					"index": "general_graylog_92",
					"node": "QTbaXWm2TbWy5eO9_oiS8g",
					"reason": {
						"type": "parse_exception",
						"reason": "failed to parse date field [2020-04-07H00:00:00.000] with format [yyyy-MM-dd HH:mm:ss.SSS]",
						"caused_by": {
							"type": "illegal_argument_exception",
							"reason": "Parse failure at index [10] of [2020-04-07H00:00:00.000]"
						}
					}
				}
			]
		},
		"status": 400
	}`
)

func MockHandler(w http.ResponseWriter, r *http.Request) {

	body, _ := ioutil.ReadAll(r.Body)

	from := gjson.GetBytes(body, `aggregations.range_aggs.range.ranges.0.from`).String()
	to := gjson.GetBytes(body, `aggregations.range_aggs.range.ranges.0.to`).String()

	if !strings.Contains(r.URL.String(), "test_index") {
		w.WriteHeader(400)
		return
	}

	_, err := time.Parse("2006-02-01 15:04:05.000", from)
	if err != nil {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write([]byte(ESWrongTimeFormat))
		if err != nil {
			fmt.Printf("error write byte err: %v", err)
		}
		return
	}

	_, err = time.Parse("2006-02-01 15:04:05.000", to)
	if err != nil {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write([]byte(ESWrongTimeFormat))
		if err != nil {
			fmt.Printf("error write byte err: %v", err)
		}
		return
	}

	// A very simple health check.
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte(ESSuccessResult1))
	if err != nil {
		fmt.Printf("error write byte err: %v", err)
	}

}

func ESMock(mockserver *httptest.Server) (*elastic.Client, error) {

	client, err := elastic.NewClient(
		elastic.SetHttpClient(mockserver.Client()),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetURL(mockserver.URL),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func TestGetLatencyPerStatus(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dateFormat := "2006-01-02"
	dateTimeFormat := "2006-01-02 15:04:05.000"

	now, _ := time.Parse(dateFormat, time.Now().Format(dateFormat))
	startTime := now.AddDate(0, 0, -1).Format(dateTimeFormat)
	endTime := now.AddDate(0, 0, -1).Add(24*time.Hour - 1*time.Microsecond).Format(dateTimeFormat)
	// indexName := "test_index"
	statusCode := "2"
	// servicesMaxNum := 500

	esResult := elastic.SearchResult{
		Aggregations: elastic.Aggregations{},
	}

	Convey("Given we have new day to get", t, func() {

		es := storesmock.NewMockElasticSearch(ctrl)
		db := storesmock.NewMockMySQL(ctrl)
		wg := sync.WaitGroup{}
		// loc, _ := time.LoadLocation("UTC")

		module := NewModule(&ModuleParam{
			ES:    es,
			MSQL:  db,
			WG:    &wg,
			Start: &startTime,
			End:   &endTime,
		})

		Convey("logging status code", func() {
			es.EXPECT().ExecuteES(&startTime, &endTime, &statusCode).Return(&esResult, nil)
			db.EXPECT().InsertToDB(gomock.Any())

			module.wg.Add(1)
			module.GetLatencyPerStatus(statusCode)
		})

		Convey("when es return error", func() {
			es.EXPECT().ExecuteES(&startTime, &endTime, &statusCode).Return(nil, errors.New("es error"))

			module.wg.Add(1)
			module.GetLatencyPerStatus(statusCode)
		})

		Convey("when db return error", func() {
			es.EXPECT().ExecuteES(&startTime, &endTime, &statusCode).Return(&esResult, nil)
			db.EXPECT().InsertToDB(gomock.Any()).Return(errors.New("db error"))

			module.wg.Add(1)
			module.GetLatencyPerStatus(statusCode)
		})

	})
}
