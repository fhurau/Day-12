package connection

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

var Con *pgx.Conn

func DatabaseConnect() {

	databaseUrl := "postgres://postgres:sudenthack@localhost:5432/personal_web_b40"

	var err error
	Con, err = pgx.Connect(context.Background(), databaseUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Success connect to database")
}
