package persister

type result struct {
	eventId string
	err error
	ok bool
}

func successResult(ID string) *result {
	return &result{
		eventId: ID,
		err: nil,
		ok: true,
	}
}

func failedResult(ID string, err error) *result {
	return &result{
		eventId: ID,
		err: err,
		ok: false,
	}
}

func (r result) ID() string {
	return r.eventId
}

func (r result) Err() error {
	return r.err
}

func (r result) Ok() bool {
	return r.ok
}

