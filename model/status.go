package model

import "github.com/pkg/errors"

type Status int

var ErrIncorrectStatusString = errors.New("incorrect status string")
var ErrIncorrectStatusCode = errors.New("incorrect status code")

const (
	Dynamic Status = 0
	Pending Status = 1
	Processing Status = 2
	Retrying Status = 3
	PartialSuccess Status = 4
	Success Status = 5
	Failed Status = 6
	Incorrect Status = 7
)

var statusMap = map[string]Status{
	"Dynamic": Dynamic,
	"Pending": Pending,
	"Processing": Processing,
	"Retrying": Retrying,
	"PartialSuccess": PartialSuccess,
	"Success": Success,
	"Failed": Failed,
}

func MapStringToStatus(status string) (Status, error) {
	if v, ok := statusMap[status]; ok {
		return v, nil
	}

	return Incorrect, errors.Wrapf(ErrIncorrectStatusString, "%s", status)
}

func MapStatusToString(status Status) (string, error) {
	for k, v := range statusMap {
		if v == status {
			return k, nil
		}
	}

	return "", errors.Wrapf(ErrIncorrectStatusCode, "%#v", status)
}