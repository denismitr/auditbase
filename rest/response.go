package rest

type statusMessage map[string]string

func respondAccepted() (int, statusMessage) {
	return 202, map[string]string{
		"status": "Accepted",
	}
}
