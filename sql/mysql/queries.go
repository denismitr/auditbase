package mysql

const createEvent = `
	INSERT INTO events (
		id, parent_event_id, actor_id, 
		actor_type_id, actor_service_id, target_id, 
		target_type_id, target_service_id, event_name,
		emitted_at, registered_at, delta
	) VALUES (
		UUID_TO_BIN(:id), UUID_TO_BIN(:parent_event_id), :actor_id, 
		UUID_TO_BIN(:actor_type_id), UUID_TO_BIN(:actor_service_id), :target_id, 
		UUID_TO_BIN(:target_type_id), UUID_TO_BIN(:target_service_id), :event_name, 
		:emitted_at, :registered_at, :delta
	)
`

const selectEvents = `
	SELECT 
		BIN_TO_UUID(e.id) as id, BIN_TO_UUID(parent_event_id) as parent_event_id,
		actor_id, BIN_TO_UUID(actor_type_id) as actor_type_id, 
		BIN_TO_UUID(actor_service_id) as actor_service_id, 
		target_id, BIN_TO_UUID(target_type_id) as target_type_id, 
		BIN_TO_UUID(target_service_id) as target_service_id, 
		event_name, emitted_at, registered_at, delta,
		ams.name as actor_service_name, tms.name as target_service_name,
		ams.description as actor_service_description, tms.description as target_service_description,
		at.name as actor_type_name, at.description as actor_type_description,
		tt.name as target_type_name, tt.description as target_type_description 
	FROM events as e
		INNER JOIN microservices as ams
	ON ams.id = e.actor_service_id
		INNER JOIN microservices as tms
	ON tms.id = e.target_service_id
		INNER JOIN actor_types as at
	ON at.id = e.actor_type_id
		INNER JOIN target_types as tt
	ON tt.id = e.target_type_id
`
