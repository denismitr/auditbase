-- init random
math.randomseed(os.time())

targetTypes = {'article','subscription','comment','foo-target','bar-target'}
services = {'back-office','front','billing','auth-service','security-service'}
actorTypes = {'user','admin','cto','promoter','employee','editor','hacker'}
eventNames = {'SOMETHING_PUBLISHED','SOMETHING_HAPPENED','SOMETHING_CRASHED','SOMETHING_DELETED'}

-- the request function that will run at each request
request = function() 
    url =  '/api/v1/events'

    actorId = '' .. math.random( 1, 20000 )
    targetId = '' .. math.random( 1, 20000 )
    emittedAt = math.random(1000000000, 2000000000)
    targetTypeName = targetTypes[math.random( #targetTypes )]
    targetServiceName = services[math.random( #services )]
    actorTypeName = actorTypes[math.random( #actorTypes )]
    actorServiceName = services[math.random( #services )]
    eventName = eventNames[math.random( #eventNames )]

    body = string.format(
        '{"targetId":"%s","targetType":{"name":"%s"},"targetService":{"name":"%s"},"actorId":"%s","actorType":{"name":"%s"},"actorService":{"name":"%s"},"eventName":"%s","emittedAt":%d,"delta":{"status":["PUBLISHED","UNPUBLISHED"],"foo":[1,2],"baz":["abcdef","fedcba"]}}',
        targetId, targetTypeName, targetServiceName, actorId, actorTypeName,
        actorServiceName, eventName, emittedAt)

    return wrk.format('POST', url, {['Content-Type'] = 'application/json', ['Accept'] = 'application/json'}, body) 
end