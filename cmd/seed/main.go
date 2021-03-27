package main

import (
	"flag"
	"github.com/denismitr/auditbase/internal/model"
	"github.com/denismitr/auditbase/internal/utils/seeder"
	"log"
	"os"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var endpoint string
	flag.StringVar(&endpoint, "endpoint", "localhost:3000/api/v1/actions", "Endpoint of POST and PATCH controllers")
	flag.Parse()

	lg := log.New(os.Stderr, "Actions Seeder ", log.LstdFlags)

	errCh := make(chan error)
	create := seeder.GenerateNewActions(1550, model.CreateAction)
	various := seeder.GenerateNewActions(2000, model.AnyAction)
	del := seeder.GenerateNewActions(9000, model.DeleteAction)
	update := seeder.GenerateNewActions(3000, model.UpdateAction)

	sender := seeder.NewSender(endpoint, lg)

	lg.Println("Starting to dispatch new actions")
	resultCh := sender.DispatchNewActions(errCh, create, various, del, update)

	for {
		select {
		case err, ok := <-errCh:
			if ok {
				lg.Println(err)
				return err
			} else {
				lg.Println("Error ch closed. Exiting")
				return nil
			}
		case result, ok := <-resultCh:
			if ok {
				lg.Printf("%#v", result)
			} else {
				lg.Println("Result ch closed. Exiting")
				return nil
			}
		}
	}
}
