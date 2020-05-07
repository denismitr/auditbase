package errtype

type StringError string

func (se StringError) Error() string {
	return string(se)
}
