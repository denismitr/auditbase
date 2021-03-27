package rest

type itemResource struct {
	Status string      `json:"status,omitempty"`
	Data   interface{} `json:"data"`
}

type collectionResource struct {
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta"`
}

type inspectResource struct {
	ConnectionStatus string `json:"connectionStatus"`
	Messages         int    `json:"messages"`
	Consumers        int    `json:"consumers"`
}
