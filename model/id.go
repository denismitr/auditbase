package model

type ID uint64

func (id ID) Value() uint64 {
	return uint64(id)
}