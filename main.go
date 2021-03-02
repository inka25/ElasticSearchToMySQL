package main

import (
	"fmt"
	"log"
	"os"
	"service-log/dialectutils"
	"service-log/internal/endpoint"
	"service-log/stores/elasticsearch"
	"service-log/stores/msql"
	"time"

	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/elastic/v6"
	"github.com/robfig/cron/v3"
	"gopkg.in/ini.v1"
)

type Config struct {
	CronSpecDaily            string `ini:"cron_spec"`
	SQLPass                  string `ini:"sql_user_pass"`
	SQLProtocol              string `ini:"sql_protocol"`
	SQLAddress               string `ini:"sql_address"`
	SQLMaxIdleConn           int    `ini:"sql_max_idle_conn"`
	SQLMaxOpenConn           int    `ini:"sql_max_open_conn"`
	SQLConnMaxLifetimeSecond int    `ini:"sql_conn_max_lifetime_second"`
	DBName                   string `ini:"sql_database"`
	TableName                string `ini:"sql_tablename"`
	ESAddress                string `ini:"es_address"`
	IndexName                string `ini:"es_indexname"`
	Location                 string `ini:"time_location"`
	ServicesMaxNum           int    `ini:"service_max_number"`
	Env                      string `ini:"env"`
}

type LogWriter struct {
	Loc time.Location
	Env string
}

func (l LogWriter) Write(bytes []byte) (int, error) {
	return fmt.Printf("%s [service-performance-%s] %s", time.Now().In(&l.Loc).Format("2006-01-02 15:04:05"), l.Env, string(bytes))
}

func main() {
	//read .ini file
	config, err := LoadConfig()
	if err != nil {
		log.Printf("error to load config, err: %v", err)
		return
	}

	loc, err := time.LoadLocation(config.Location)
	if err != nil {
		log.Printf("timezone parsing error, %v", err)
		return
	}

	ServiceLogger := LogWriter{
		Loc: *loc,
		Env: config.Env,
	}
	log.SetOutput(ServiceLogger)
	log.SetFlags(0)

	db, err := sqlx.Connect("mysql", fmt.Sprintf("%s@%s(%s)/%s",
		config.SQLPass,
		config.SQLProtocol,
		config.SQLAddress,
		config.DBName))
	if err != nil {
		log.Printf("failed to connect to mysql, err: %v", err)
		return
	}
	db.SetConnMaxLifetime(time.Duration(config.SQLConnMaxLifetimeSecond) * time.Second)
	db.SetMaxIdleConns(config.SQLMaxIdleConn)
	db.SetMaxOpenConns(config.SQLMaxOpenConn)
	defer db.Close()

	dialectutils.PrepareQueryBuilder("mysql")

	es, err := elastic.NewClient(
		elastic.SetURL(config.ESAddress),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetErrorLog(
			log.New(&ServiceLogger, "", 0),
		),
	)
	if err != nil {
		log.Printf("failed to connect to elasticsearch, err: %v", err)
		return
	}

	start, end := LoadTime(loc)

	esService := elasticsearch.New(es, &config.IndexName, &config.ServicesMaxNum)
	sqlService := msql.New(db, &config.TableName)

	dailyJob := endpoint.NewModule(&endpoint.ModuleParam{
		ES:    esService,
		MSQL:  sqlService,
		Start: start,
		End:   end,
	},
	)

	crn := cron.New(cron.WithLocation(loc))
	_, err = crn.AddJob(config.CronSpecDaily, dailyJob)
	if err != nil {
		log.Printf("cron failed, err: %v", err)
		return
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGSTOP)

	log.Println("starting cron...")
	crn.Start()

	<-ch
	crn.Stop()
	db.Close()
	log.Println("shutting down")
	os.Exit(0)
}

func LoadConfig() (Config, error) {
	config := Config{}
	confFile, err := os.Open("./conf/servicelog.ini")
	if err != nil {
		return config, err
	}
	defer confFile.Close()

	err = ini.MapTo(&config, confFile)
	if err != nil {
		return config, err
	}
	return config, nil
}

func LoadTime(loc *time.Location) (start, end *string) {

	dateFormat := "2006-01-02"
	dateTimeFormat := "2006-01-02 15:04:05.000"

	now, _ := time.Parse(dateFormat, time.Now().In(loc).Format(dateFormat))
	startstr := now.AddDate(0, 0, -1).Format(dateTimeFormat)
	endstr := now.AddDate(0, 0, -1).Add(24*time.Hour - 1*time.Microsecond).Format(dateTimeFormat)

	return &startstr, &endstr
}
