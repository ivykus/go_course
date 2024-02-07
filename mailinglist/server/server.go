package main

import (
	"database/sql"
	"log"
	"sync"

	"github.com/alexflint/go-arg"
	"github.com/ivykus/gocourse/mailinglist/grpcapi"
	"github.com/ivykus/gocourse/mailinglist/jsonapi"
	"github.com/ivykus/gocourse/mailinglist/mdb"
)

var args struct {
	DbPath   string `arg:"env:MAILINGLIST_DB"`
	BindJson string `arg:"env:MAILINGLIST_BIND_JSON"`
	BindGrpc string `arg:"env:MAILINGLIST_BIND_GRPC"`
}

func main() {
	arg.MustParse(&args)

	if args.DbPath == "" {
		args.DbPath = "list.db"
	}

	if args.BindJson == "" {
		args.BindJson = ":8080"
	}

	if args.BindGrpc == "" {
		args.BindGrpc = ":8081"
	}

	log.Printf("using database '%v'\n", args.DbPath)
	db, err := sql.Open("sqlite3", args.DbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mdb.TryCreate(db)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		log.Println("Starting JSON API server...")
		jsonapi.Serve(db, args.BindJson)
		wg.Done()
	}()

	wg.Add(1)

	go func() {
		log.Println("Starting gRPC API server...")
		grpcapi.Serve(db, args.BindGrpc)
		wg.Done()
	}()

	wg.Wait()
}
