package env

import (
	"github.com/joho/godotenv"
)

func LoadFromDotEnv(files ...string) {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
}
