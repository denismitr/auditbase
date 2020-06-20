# Auditbase 
### an audit system for domain events, specifically designed for microservices

### WIP

Auditbase is an application that collects events from your distributed system. 
Denormalizes events data to make them better suitable for analytics

consists of **receiver** REST API, **back-office** REST API, 
AMQP events consumer, AMQP errors consumer

at the moment only MySQL storage is available, CockroachDB and MongoDB are planned.
Redis is used as cache for faster denormalization, RabbitMQ as message broker

### RUN DEV MODE
```make up```

### UNIT TESTS
RUN:

- ```make mock```
- ```make test```

### RUN "WRK" BENCHMARK
install `wrk` tool
```
wrk -c5 -t3 -R300 -d166s -s ./test/lua/events.lua --latency http://localhost:8888
```

## REST API

### RECEIVER API
Receives events to put them into queue for later processing by consumers

-  POST /api/v1/events


### BACK-OFFICE API
API suitable for a back-office admin panel

##### Events
-  GET /api/v1/events
-  GET /api/v1/events/:id
-  GET /api/v1/events/count
-  GET /api/v1/events/queue
-  DELETE /api/v1/events/:id

##### Microservices
- GET /api/v1/microservices
- POST /api/v1/microservices
- GET /api/v1/microservices/:id
- PUT /api/v1/microservices/:id

##### Entities
- GET /api/v1/entities
- GET /api/v1/entities/:id

##### Properties
- GET /api/v1/properties
- GET /api/v1/properties/:id

##### Changes
- GET /api/v1/changes
- GET /api/v1/changes/:id

## TODO
- unit tests
- more integration tests
- SQL query builder for all queries
- refactor ID to object where suitable
- create back-office UI dashboard
- MongoDB as alternative storage
- research GRPC and Protobuf as alternative to HTTP REST
- research NATS as alternative to RabbitMQ

