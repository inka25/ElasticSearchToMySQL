package msql

import (
	"github.com/doug-martin/goqu/v9"
)

func (m *MSQL) InsertToDB(values ...[]interface{}) error {
	if len(values) == 0 {
		return nil
	}

	query, params, err := goqu.Dialect(m.client.DriverName()).Insert(goqu.T(*m.tableName)).
		Cols(
			goqu.C("service_name"),
			goqu.C("timestamp"),
			goqu.C("status_code"),
			goqu.C("doc_count"),
			goqu.C("sum"),
			goqu.C("avg"),
			goqu.C("percentiles"),
		).
		Vals(values...).OnConflict(
		goqu.DoUpdate("key",
			goqu.Record{
				"doc_count":   goqu.Func("VALUES", goqu.C("doc_count")),
				"sum":         goqu.Func("VALUES", goqu.C("sum")),
				"avg":         goqu.Func("VALUES", goqu.C("avg")),
				"percentiles": goqu.Func("VALUES", goqu.C("percentiles")),
				"timestamp":   goqu.Func("VALUES", goqu.C("timestamp")),
			},
		)).Prepared(true).ToSQL()
	if err != nil {
		return err
	}

	_, err = m.client.Exec(query, params...)
	return err
}
