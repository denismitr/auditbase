package utils

import (
	uuid "github.com/satori/go.uuid"
	"strings"
)

// UUID4Generator - new UUID4 generator
type UUID4Generator interface {
	Generate() string
}

// SatoriUUID4Generator - generates UUIDs using satori package
type SatoriUUID4Generator struct{}

// Generate UUID V4 String
func (g *SatoriUUID4Generator) Generate() string {
	return strings.ReplaceAll(uuid.NewV4().String(), "-", "")
}

// NewUUID4Generator - creates new UUID4Generator
func NewUUID4Generator() UUID4Generator {
	return &SatoriUUID4Generator{}
}