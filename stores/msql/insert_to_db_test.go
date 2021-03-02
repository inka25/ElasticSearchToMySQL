package msql

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	. "github.com/smartystreets/goconvey/convey"
)

func TestInsertToDB(t *testing.T) {

	tableName := "test_table"
	result := []interface{}{
		"service_name",
		"date",
		"status_code",
		0,
		0,
		0.0,
		"eyIxLjAiOjAsIjI1LjAiOjEsIjUuMCI6MCwiNTAuMCI6MTEsIjc1LjAiOjI0LjgxMTQ1ODUyMTczOTgxNiwiOTUuMCI6OTYuNDQ1NzQ1MTk2ODUxNzUsIjk5LjAiOjE5OC45NDYzOTU5NzgwMTYyfQ==",
	}

	Convey("Recording new result to store", t, func() {

		db, dbMock, err := sqlmock.New()
		So(err, ShouldBeNil)
		defer db.Close()

		dbx := sqlx.NewDb(db, "mysql")
		msql := New(dbx, &tableName)

		Convey("When succesfully recorded to db", func() {
			dbMock.ExpectExec("INSERT INTO \"test_table\"").WillReturnResult(sqlmock.NewResult(1, 1))
			err := msql.InsertToDB([]interface{}{result})
			So(err, ShouldBeNil)
		})

		Convey("When there is an error when recording to db", func() {
			dbMock.ExpectExec("INSERT INTO \"test_table\"").WillReturnError(errors.New("error db"))
			err := msql.InsertToDB([]interface{}{result})
			So(err, ShouldNotBeNil)
		})

	})

}
