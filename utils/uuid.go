package utils

import uuid "github.com/satori/go.uuid"

// UUID4Generatgor - new UUID4 generator
type UUID4Generatgor interface {
	Generate() string
}

// SatoriUUID4Generator - generates UUIDs using satori package
type SatoriUUID4Generator struct{}

// Generate UUID V4 String
func (g *SatoriUUID4Generator) Generate() string {
	return uuid.NewV4().String()
}

// NewUUID4Generator - creates new UUID4Generatgor
func NewUUID4Generator() UUID4Generatgor {
	return &SatoriUUID4Generator{}
}
