package endpoint

import (
	"encoding/json"
	"log"
	"service-log/internal/extractor"
)

func (m *Module) GetLatencyPerStatus(statusCode string) {
	defer m.wg.Done()

	strCod := statusCode
	if len(statusCode) == 1 && statusCode != "0" {
		strCod = statusCode + "xx"
	}
	log.Printf("getting log with status %s \n", strCod)

	result, err := m.es.ExecuteES(m.startTime, m.endTime, &statusCode)
	if err != nil {
		log.Printf("failed to get from elasticsearch, err: %v\n", err)
		return
	}

	bt, _ := json.MarshalIndent((*result).Aggregations, "", " ")
	values := extractor.ExtractValues(bt, m.startTime, &statusCode)

	err = m.msql.InsertToDB(values...)
	if err != nil {
		log.Printf("failed to write to db, err: %v\n", err)
		return
	}

	log.Printf("recorded %d services with status %s to db\n", len(values), strCod)
}
