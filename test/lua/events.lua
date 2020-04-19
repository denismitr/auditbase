-- init random
math.randomseed(os.time())

targetEntities = {'article','subscription','comment','foo-target','bar-target','baz-target','tariff','promocode'}
services = {'back-office','front','billing','auth-service','security-service','foo-service','bar-service','baz-service'}
actorEntities = {'user','admin','cto','promoter','employee','editor','hacker','foo-actor','bar-actor','baz-actor'}
eventNames = {'SOMETHING_PUBLISHED','SOMETHING_HAPPENED','SOMETHING_CRASHED','SOMETHING_DELETED','SOMETHING_FOO','SOMETHING_BAR'}

-- the request function that will run at each request
request = function() 
    url =  '/api/v1/events'

    actorId = '' .. math.random( 1, 20000 )
    targetId = '' .. math.random( 1, 20000 )
    emittedAt = math.random(1000000000, 2000000000)
    targetEntity = targetTypes[math.random( #targetTypes )]
    targetService = services[math.random( #services )]
    actorEntity = actorTypes[math.random( #actorTypes )]
    actorService = services[math.random( #services )]
    eventName = eventNames[math.random( #eventNames )]

    body = string.format(
        '{"targetId":"%s","targetEntity":%s","targetService":"%s","actorId":"%s","actorEntity":"%s","actorService":"%s","eventName":"%s","emittedAt":%d,"delta":{"status":["PUBLISHED","UNPUBLISHED"],"foo":[1,2],"baz":["abcdef","fedcba"],"bar":[null,"bar-state"]}}',
        targetId, targetEntity, targetService, actorId, actorEntity,
        actorService, eventName, emittedAt)

    return wrk.format('POST', url, {['Content-Type'] = 'application/json', ['Accept'] = 'application/json'}, body) 
end