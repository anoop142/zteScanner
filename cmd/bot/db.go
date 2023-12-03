package main

import(
	"database/sql"
	"context"
	"time"
	"os"

	_ "modernc.org/sqlite"
)

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		return nil, err
	}

	return db, err
}


// create db and tables
func initDB(dbPath string)error{
	f,err := os.Create(dbPath)
	if err != nil{
		return err
	}
	defer f.Close()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil{
		return err
	}


	if err!= nil{
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		return err
	}

	// create known_devices table
	query1 := `CREATE TABLE known_devices (
		  id INT,
		  name TEXT NOT NULL,
		  mac TEXT NOT NULL,
		  ignore INT)`

	_, err = db.ExecContext(ctx, query1)
	if err != nil{
		return err
	}

	query2 := `CREATE TABLE ignore_devices (
	          id INT,
		  name TEXT NOT NULL,
		  mac TEXT NOT NULL)`
	_, err = db.ExecContext(ctx, query2)
	if err!= nil{
		return err
	}

	return nil
}
	


