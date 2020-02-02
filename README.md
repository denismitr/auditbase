# Auditbase - an audit log system suitable for microservices

### RUN
```make up```

### RUN "WRK" BENCHMARK
```
wrk -c5 -t3 -R300 -d166s -s ./test/lua/events.lua --latency http://localhost:8888
```

## TODO
- graceful shutdown
- healthcheck for rest API
- more tests

### ENDPOINTS

##### Events
-  POST /api/v1/events
-  GET /api/v1/events
-  GET /api/v1/events/count
-  GET /api/v1/events/queue
-  DELETE /api/v1/events/:id
-  GET /api/v1/events/:id

##### Microservices
- /api/v1/microservices
- /api/v1/microservices
- /api/v1/microservices/:id
- /api/v1/microservices/:id

##### Actor types
- /api/v1/actor-types
- /api/v1/actor-types/:id

##### Target types
- /api/v1/target-types
- /api/v1/target-types/:id