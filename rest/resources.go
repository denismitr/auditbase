package rest

type resourceSerializer interface {
	ToJSON() respource
}

type respource struct {
	Data interface{} `json:"data"`
}

type inspectResource struct {
	Messages  int `json:"messages"`
	Consumers int `json:"consumers"`
}

func (ir inspectResource) ToJSON() respource {
	return respource{Data: ir}
}
