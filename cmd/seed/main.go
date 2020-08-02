package main

import (
	"fmt"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/test/seeder"
)

func main() {
	create := seeder.New(150, false, model.Create).Seed()
	unknown := seeder.New(200, false, model.Unknown).Seed()
	del := seeder.New(90, false, model.Delete).Seed()
	update := seeder.New(300, false, model.Update).Seed()

	results := seeder.Send("http://localhost:8888/api/v1/events", create, unknown, del, update) // fixme
	for result := range results {
		fmt.Printf("%#v", result)
	}
}
