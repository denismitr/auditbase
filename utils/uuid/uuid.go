package uuid

import satoriUuid "github.com/satori/go.uuid"

// UUID4Generator - new UUID4 generator
type UUID4Generator interface {
	Generate() string
}

// SatoriUUID4Generator - generates UUIDs using satori package
type SatoriUUID4Generator struct{}

// Generate UUID V4 String
func (g *SatoriUUID4Generator) Generate() string {
	return satoriUuid.NewV4().String()
}

// NewUUID4Generator - creates new UUID4Generatgor
func NewUUID4Generator() UUID4Generator {
	return &SatoriUUID4Generator{}
}
