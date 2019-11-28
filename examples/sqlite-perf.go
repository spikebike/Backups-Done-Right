package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
)

var (
	inserts   int
	chunks    int
	PerCommit int
)

const DataBaseName = "./foo.db"

func run() {
	var tx *sql.Tx

	os.Remove(DataBaseName)

	db, err := sql.Open("sqlite3", DataBaseName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	_, err = db.Exec("create table foo (id INT)")
	if err != nil {
		fmt.Printf("%q: %s\n", err)
		return
	}

	tx, err = db.Begin()
	if err != nil {
		fmt.Println(err)
	}

	chunks := inserts / PerCommit

	for a := 0; a <= chunks; a++ {
		tx, err = db.Begin()
		if err != nil {
			fmt.Println(err)
		}
		for i := 0; i <= PerCommit; i++ {
			stmt, err := tx.Prepare("insert into foo(id) values(?)")
			if err != nil {
				fmt.Println(err)
				return
			}
			defer stmt.Close()
			_, err = stmt.Exec(i)
			if err != nil {
				fmt.Println(err)
				return
			}

		}
		tx.Commit()
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: sqlite-performacetest <number of inserts> <inserts per commit>\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		usage()
	}
	fmt.Sscanf(args[0], "%d", &inserts)
	fmt.Sscanf(args[1], "%d", &PerCommit)
	fmt.Printf("\nStarting inserts = %d PerCommit=%d\n", inserts, PerCommit)
	t0 := time.Now().UnixNano()
	run()
	t1 := time.Now().UnixNano()
	duration := float64(t1-t0) / 1000000000
	fmt.Printf("%d inserts, %3d inserts/commit in %4.1f seconds for %8.2f inserts/sec\n", inserts, PerCommit, duration, float64(inserts)/duration)

}
