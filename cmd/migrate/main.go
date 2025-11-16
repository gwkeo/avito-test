package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
)

const (
	dialect = "postgres"
)

var (
	flags = flag.NewFlagSet("migrate", flag.ExitOnError)
	dir   = flags.String("dir", "./migrations", "path to migrations")
)

func main() {

	// Так как у нас один параметр -dir, его и передаем в Parse
	flags.Parse(os.Args[1:])
	args := flags.Args()
	if len(args) < 2 {
		usage("not enough arguments")
		return
	}

	dbString, command := args[0], args[1]
	if err := goose.SetDialect(dialect); err != nil {
		log.Fatalf("unable to set dialect: %v", err)
		return
	}
	db, err := sql.Open("pgx", dbString)
	if err != nil {
		log.Fatalf("unable to open DB: %v", err)
	}
	defer db.Close()

	if err = goose.Run(command, db, *dir, args[:2]...); err != nil {
		log.Fatalf("unable to run command: %v", err)
	}
}

func usage(msg string) {
	fmt.Println(msg)
	fmt.Println("migrate [-dir] <dbString> <command> <args...>")
	flags.Usage()
}
