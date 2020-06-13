package model

type RepositoryFactory interface {
	Properties() PropertyRepository
	Events() EventRepository
	Entities() EntityRepository
	Microservices() MicroserviceRepository
}
