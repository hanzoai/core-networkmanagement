package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func openDB(host string, dbUser string, dbPort string, dbTable string, dbCertDir string) {
	dbOpts := "?sslmode=disable"
	if dbCertDir != "" {
		dbOpts = "?ssl=true&sslmode=require&sslrootcert=" + dbCertDir + "/ca.crt&sslkey=" + dbCertDir + "/client." + dbUser + ".key&sslcert=" + dbCertDir + "/client." + dbUser + ".crt"
	}
	dbcreds := fmt.Sprintf("postgresql://%s@%s:%s/%s%s",
		dbUser, host, dbPort, dbTable, dbOpts)
	var dberr error
	db, dberr = sqlx.Connect("pgx", dbcreds)
	if dberr != nil {
		log.Fatal(dberr)
	}
	initDB(db, dbTable)
}
func initDB(db *sqlx.DB, database string) {
	tables := [5]string{"Users", "Networks", "Hosts", "Groups", "Domains"}
	checkInit, err := db.Prepare("select count(*) as count from information_schema.tables where table_name = $1;")
	if err != nil {
		log.Fatal(err)
	}
	initialized := true
	count := 0
	for _, table := range tables {
		err := checkInit.QueryRow(table).Scan(&count)
		initialized = initialized && (count == 1)
		if err != nil {
			log.Fatal(err)
		}
	}
	if !initialized {
		fmt.Print("Initializing database")
		file, ferr := ioutil.ReadFile("init.sql")
		if ferr != nil {
			log.Fatal(ferr)
		}
		requests := strings.Split(string(file), ";\n")
		for _, request := range requests {
			_, err := db.Exec(request)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
