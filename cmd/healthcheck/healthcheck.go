package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	_, err := http.Get(fmt.Sprintf("http://127.0.0.1:%s/health", os.Getenv("HEALTH_PORT")))
	if err != nil {
		log.Printf("\nhealthcheck failed on port %s because: %s", os.Getenv("HEALTH_PORT"), err.Error())
		os.Exit(1)
	}
}
