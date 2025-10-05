package main

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Area struct {
	ID      int    `db:"id"`
	Name    string `db:"name"`
	Polygon string `db:"polygon"`
}

func main() {
	dsn := "host=localhost port=5432 user=altrinity password=altrinity dbname=geodb sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	var areas []Area
	err = db.Select(&areas, "SELECT id, name, ST_AsText(polygon) AS polygon FROM areas")
	if err != nil {
		log.Fatalln(err)
	}

	for _, a := range areas {
		log.Printf("%d: %s", a.ID, a.Name)
	}
}
