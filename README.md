# Auditbase 
### an audit log system specifically designed for microservices

### WIP

Auditbase denormalizes events to make them suitable for analytics

consists of **receiver** REST API, **back-office** REST API, 
RabbitMQ events consumer, RabbitMQ errors consumer

at the moment only MySQL storage is available, MongoDB is planned,
Redis as cache for faster denormalization, RabbitMQ as message broker

### RUN DEV MODE
```make up```

### TESTS
RUN:

- ```make mock```
- ```make test```

### RUN "WRK" BENCHMARK
install `wrk` tool
```
wrk -c5 -t3 -R300 -d166s -s ./test/lua/events.lua --latency http://localhost:8888
```

## REST API

### RECEIVER ENDPOINT
-  POST /api/v1/events

### BACK-OFFICE ENDPOINTS

##### Events
-  GET /api/v1/events
-  GET /api/v1/events/:id
-  GET /api/v1/events/count
-  GET /api/v1/events/queue
-  DELETE /api/v1/events/:id

##### Microservices
- /api/v1/microservices
- /api/v1/microservices
- /api/v1/microservices/:id
- /api/v1/microservices/:id

##### Entities
- /api/v1/entities
- /api/v1/entities/:id

## TODO
- unit tests
- more integration tests
- SQL query builder
- refactor ID to object where suitable
- create back-office UI dashboard
- MongoDB as alternative storage
- research GRPC

