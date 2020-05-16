-- init random
math.randomseed(os.time())

targetEntities = {'article','subscription','comment','foo-target','bar-target','baz-target','tariff','promocode'}
services = {'back-office','front','billing','auth-service','security-service','foo-service','bar-service','baz-service'}
actorEntities = {'user','admin','cto','promoter','employee','editor','hacker','foo-actor','bar-actor','baz-actor'}
eventNames = {'SOMETHING_PUBLISHED','SOMETHING_HAPPENED','SOMETHING_CRASHED','SOMETHING_DELETED','SOMETHING_FOO','SOMETHING_BAR'}
properties = {'foo-property', 'bar-property', 'baz-property', 'title', 'name', 'number', 'text', 'body', 'date'}
stringChanges = {'null', '"boo"', '"foo"', '"baz"', '"abc"', '"123abc"'}
numericChanges = {'null', '123', '987542', '1235.973', '999', '0'}
dateChanges = {'null', '"1999-10-12"', '"2000-10-01"', '"1890-10-10"'}
enumChanges = {'null', '"ACTIVE"', '"DELETED"', '"ON"', '"OFF"', '"MODERATED"', '"BANNED"', '"FAILED"'}

backOffice = {
    name = "back-office",
    entities = {'user','admin','cto','promoter'}
}

front = {
    name = "front-service",
    entities = {'token','ga','ya'}
}

auth = {
    name = "auth-service",
    entities = {'user','mailing-list','hashing-alg'}
}

billing = {
    name = "billing-service",
    entities = {'subscription','tariff','plan','promocode'}
}

targetServices = {backOffice, front, auth, billing}

-- the request function that will run at each request
request = function() 
    url =  '/api/v1/events'

    actorId = '' .. math.random( 1, 20000 )
    targetId = '' .. math.random( 1, 20000 )
    emittedAt = math.random(1000000000, 2000000000)
    targetService = targetServices[ math.random( #targetServices ) ]
    targetEntity = targetService.entities[ math.random(#targetService.entities) ]
    actorEntity = actorEntities[math.random( #actorEntities )]
    actorService = services[math.random( #services )]
    eventName = eventNames[math.random( #eventNames )]

    delta = createDelta()

    body = string.format(
        '{"targetId":"%s","targetEntity":"%s","targetService":"%s","actorId":"%s","actorEntity":"%s","actorService":"%s","eventName":"%s","emittedAt":%d,"delta":%s}',
        targetId, targetEntity, targetService, actorId, actorEntity,
        actorService, eventName, emittedAt, delta)

    print(body)

    return wrk.format('POST', url, {['Content-Type'] = 'application/json', ['Accept'] = 'application/json'}, body) 
end

createDelta = function()
    enumProp = properties[ math.random( #properties ) ]
    stringProp = properties[ math.random( #properties ) ]
    numericProp = properties[ math.random( #properties ) ]
    dateProp = properties[ math.random( #properties ) ]

    enumChangedFrom = enumChanges[ math.random( #enumChanges ) ]
    enumChangedTo = enumChanges[ math.random( #enumChanges ) ]

    stringChangedFrom = stringChanges[ math.random( #stringChanges ) ]
    stringChangedTo = stringChanges[ math.random( #stringChanges ) ]

    numericChangedFrom = numericChanges[ math.random( #numericChanges ) ]
    numericChangedTo = numericChanges[ math.random( #numericChanges ) ]

    dateChangedFrom = dateChanges[ math.random( #dateChanges ) ]
    dateChangedTo = dateChanges[ math.random( #dateChanges ) ]

    delta = string.format(
        '{"%s":[%s,%s],"%s":[%s,%s],"%s":[%s,%s],"%s":[%s,%s]}',
        enumProp, enumChangedFrom, enumChangedTo,
        stringProp, stringChangedFrom, stringChangedTo,
        numericProp, numericChangedFrom, numericChangedTo,
        dateProp, dateChangedFrom, dateChangedTo)

    return delta
end