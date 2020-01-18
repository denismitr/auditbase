package rest

type resourceSerializer interface {
	ToJSON() response
}

type response struct {
	Data interface{} `json:"data"`
}

type inspectResource struct {
	Messages  int `json:"messages"`
	Consumers int `json:"consumers"`
}

func (ir inspectResource) ToJSON() response {
	return response{Data: ir}
}
