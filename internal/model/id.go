package model

type ID uint64

func (id ID) UInt64() uint64 {
	return uint64(id)
}

func (id ID) Int64() int64 {
	return int64(id)
}

func (id ID) Int() int {
	return int(id)
}

func (id ID) Valid() bool {
	return id > 0
}

func (id ID) Empty() bool {
	return id == 0
}

type UID string

func (uid UID) Empty() bool {
	return len(uid) == 0
}

func (uid UID) Valid() bool {
	if len(uid) != 32 {
		return false
	}

	// TODO: regexp check
	return true
}

func (uid UID) String() string {
	return string(uid)
}