package env

import (
	"os"
	"strconv"
	"strings"
)

func String(key string) string {
	return os.Getenv(key)
}

func StringOrDefault(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	return v
}

func MustString(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("no key " + key + " in environment variables")
	}

	return v
}

func Int(key string) int {
	v := os.Getenv(key)

	i, _ := strconv.Atoi(v)

	return i
}

func IntOrDefault(key string, def int) int {
	v := os.Getenv(key)

	i, err := strconv.Atoi(v)
	if err != nil || i == 0 {
		return def
	}

	return i
}

func IsTruthy(key string) bool {
	v := strings.ToLower(os.Getenv(key))

	if v == "" {
		return false
	}

	if v == "0" {
		return false
	}

	if v == "false" || v == "off" || v == "no" {
		return false
	}

	return true
}

func IsFalsy(key string) bool {
	return !IsTruthy(key)
}
