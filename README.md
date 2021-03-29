# Auditbase 
### an audit system for domain actions and events, specifically designed for microservices

### WIP

Auditbase is an application that collects actions from your distributed systems. 
Denormalizes data to make them better suitable for analysis

consists of **receiver** REST API, **back-office** REST API, 
AMQP consumer, AMQP errors consumer

at the moment only MySQL storage is available, CockroachDB and MongoDB are planned.
Redis is used as caching, RabbitMQ as message broker

### RUN DEV MODE
```make up```

### UNIT TESTS
RUN:

- ```make test```

## REST API

### RECEIVER API
Receives actions to put them into queue for later processing by consumers

-  POST /api/v1/actions
- PATCH /api/v1/actions (updates status nly)

##### Payload sample
```json
{
  "uid": "111d2edbf207452eae7ec258271ee98c",
  "parentUid": "44402edbf207452eae7ec258271ee98c",
  "targetExternalId": "9309213",
  "targetEntity": "article3",
  "targetService": "article-storage-4",
  "actorExternalId": "9",
  "actorEntity": "promoter3",
  "actorService": "back-office-4",
  "name": "articlePublished",
  "emittedAt": "2021-01-02 15:04:05",
  "isAsync": true,
  "status": 2,
  "details": {
    "delta": [
      {
        "propertyName": "text",
        "currentPropertyType": "string",
        "from": null,
        "to": "foo says bar"
      },
      {
        "propertyName": "title",
        "currentPropertyType": "string",
        "from": null,
        "to": "baz title"
      },
      {
        "propertyName": "rating",
        "currentPropertyType": "float",
        "from": null,
        "to": 1.1
      },
      {
        "propertyName": "views",
        "currentPropertyType": "integer",
        "from": null,
        "to": 1
      }
    ]
  }
}
```

## BACK-OFFICE API
API suitable for a back-office admin panel

### Actions
####  GET /api/v1/actions
##### Allowed filters:
- name="action_name"
- parentUid="uuid4-without-dashes"
- status=1
- actorEntityId=123
- targetEntityId=123

##### Cursor:
- page=1
- perPage=20

#### GET /api/v1/actions/:id
-  GET /api/v1/actions/count // TODO
-  GET /api/v1/actions/queue // TODO
-  DELETE /api/v1/actions/:id // TODO

### Microservices
- GET /api/v1/microservices
- POST /api/v1/microservices
- GET /api/v1/microservices/:id
- PUT /api/v1/microservices/:id

### Entities
- GET /api/v1/entities
- GET /api/v1/entities/:id

## TODO
- unit tests
- more integration tests
- cleaner worker
- replace squirrel for goqu everywhere
- MongoDB as alternative storage
- research GRPC and Protobuf as alternative to HTTP REST
- research NATS as alternative to RabbitMQ

