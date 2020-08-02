-- init random
math.randomseed(os.time())

enumProperties = {"updated", "banned", "moved", "reassigned" }

stringChanges = {'null', '"boo"', '"foo"', '"baz"', '"abc"', '"123abc"'}
numericChanges = {'null', '123', '987542', '1235.973', '999', '0'}
dateChanges = {'null', '"1999-10-12"', '"2000-10-01"', '"1890-10-10"'}
enumChanges = {'null', '"ACTIVE"', '"DELETED"', '"ON"', '"OFF"', '"MODERATED"', '"BANNED"', '"FAILED"'}

-- Entities

userEntity = {
    name = "user",
    properties = {"name", "phone", "email", "status", "address"},
    events = {"user-updated", "user-upgraded", "user-downgraded", "user-banned"},
}

adminEntity = {
    name = "admin",
    properties = {"username", "password", "email", "status", "rights"},
    events = {"admin-updated", "admin-banned", "admin-moved"}
}

actionEntity = {
    name = "action",
    properties = {"title", "createdAt", "data"},
    events = {"action-recorded"}
}

tokenEntity = {
    name = "token",
    properties = {"body", "issuedAt"},
    events = {"token-updated"}
}

subscriptionEntity = {
    name = "token",
    properties = {"type", "duration", "price", "period"},
    events = {"subscription-updated", "subscription-created"}
}

-- services

backOffice = {
    name = "back-office",
    entities = {userEntity, adminEntity, actionEntity}
}

front = {
    name = "front-service",
    entities = {actionEntity, tokenEntity}
}

auth = {
    name = "auth-service",
    entities = {userEntity}
}

billing = {
    name = "billing-service",
    entities = {subscriptionEntity}
}

services = {backOffice, front, auth, billing}

-- the request function that will run at each request
request = function() 
    url =  '/api/v1/events'

    actorId = '' .. math.random( 1, 20000 )
    targetId = '' .. math.random( 1, 20000 )
    emittedAt = math.random(1000000000, 2000000000)

    -- target
    targetService = services[ math.random( #services ) ]
    targetEntity = targetService.entities[ math.random(#targetService.entities) ]
    eventName = targetEntity.events[math.random( #targetEntity.events )]

     -- actor
    actorService = services[math.random( #services )]
    actorEntity = actorService.entities[math.random( #actorService.entities )]

    delta = createDelta(targetEntity)

    body = string.format(
        '{"targetId":"%s","targetEntity":"%s","targetService":"%s","actorId":"%s","actorEntity":"%s","actorService":"%s","eventName":"%s","emittedAt":%d,"delta":%s}',
        targetId, targetEntity.name, targetService.name, actorId, actorEntity.name,
        actorService.name, eventName, emittedAt, delta)

    print("\n")
    print(body)

    return wrk.format('POST', url, {['Content-Type'] = 'application/json', ['Accept'] = 'application/json'}, body) 
end

local function createChange(type, targetEntity)
    local propName = ""
    if type == "enum" then
       propName = math.random( #enumProperties )
    end

    local enumProp = targetEntity.properties[ math.random( #targetEntity.properties ) ]
    local stringProp = targetEntity.properties[ math.random( #targetEntity.properties ) ]
    local numericProp = targetEntity.properties[ math.random( #targetEntity.properties ) ]
    local dateProp = targetEntity.properties[ math.random( #targetEntity.properties ) ]

    local enumChangedFrom = enumChanges[ math.random( #enumChanges ) ]
    local enumChangedTo = enumChanges[ math.random( #enumChanges ) ]

    local stringChangedFrom = stringChanges[ math.random( #stringChanges ) ]
    local stringChangedTo = stringChanges[ math.random( #stringChanges ) ]

    local numericChangedFrom = numericChanges[ math.random( #numericChanges ) ]
    local numericChangedTo = numericChanges[ math.random( #numericChanges ) ]

    local dateChangedFrom = dateChanges[ math.random( #dateChanges ) ]
    local dateChangedTo = dateChanges[ math.random( #dateChanges ) ]

    local delta = string.format(
        '{"propertyName":"%s","from":"%s","to":"%s"}',
        propName, enumChangedFrom, enumChangedTo,
        stringProp, stringChangedFrom, stringChangedTo,
        numericProp, numericChangedFrom, numericChangedTo,
        dateProp, dateChangedFrom, dateChangedTo)

    return delta
end