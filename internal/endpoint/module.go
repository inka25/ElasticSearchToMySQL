package endpoint

import (
	"log"
	"service-log/stores"
	"sync"
)

var statusCodes = []string{"2", "3", "4", "500", "503", "504", "0"}

type Module struct {
	es        stores.ElasticSearch
	msql      stores.MySQL
	wg        *sync.WaitGroup
	startTime *string
	endTime   *string
}

type ModuleParam struct {
	ES    stores.ElasticSearch
	MSQL  stores.MySQL
	WG    *sync.WaitGroup
	Start *string
	End   *string
}

func NewModule(m *ModuleParam) *Module {
	var wg sync.WaitGroup

	return &Module{
		es:        m.ES,
		wg:        &wg,
		msql:      m.MSQL,
		startTime: m.Start,
		endTime:   m.End,
	}
}

func (m *Module) Run() {

	log.Printf("logging service performances from %v to %v", *m.startTime, *m.endTime)

	for _, v := range statusCodes {
		m.wg.Add(1)
		go m.GetLatencyPerStatus(v)
	}

	m.wg.Wait()
}
