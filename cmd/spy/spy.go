package main

import (
	"github.com/natewong1313/spy/internal/db"
	"github.com/natewong1313/spy/scrapers/greenhouse"
)

func main() {
	db, err := db.NewPool("postgres://user:password@127.0.0.1:5432/spydb?sslmode=disable")
	if err != nil {
		panic(err)
	}
	greenhouse.Start(db)

}
