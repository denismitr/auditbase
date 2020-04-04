package queue

type Message interface {
	Body() []byte
	ContentType() string
	Attempt() int
}

type JSONMessage struct {
	body    []byte
	attempt int
}

func NewJSONMessage(b []byte, attempt int) *JSONMessage {
	return &JSONMessage{
		body:    b,
		attempt: attempt,
	}
}

func (e *JSONMessage) Body() []byte {
	return e.body
}

func (e *JSONMessage) ContentType() string {
	return "application/json"
}

func (e *JSONMessage) Attempt() int {
	return e.attempt
}

type ReceivedMessage interface {
	Body() []byte
	Queue() string
	Attempt() int
	CloneToReque() Message
	Tag() uint64
}

type RabbitMQReceivedMessage struct {
	queueName string
	body      []byte
	attempt   int
	tag       uint64
}

func (m RabbitMQReceivedMessage) Queue() string {
	return m.queueName
}

func (m RabbitMQReceivedMessage) Body() []byte {
	return m.body
}

func (m RabbitMQReceivedMessage) Attempt() int {
	return m.attempt
}

func (m RabbitMQReceivedMessage) Tag() uint64 {
	return m.tag
}

func (m *RabbitMQReceivedMessage) CloneToReque() Message {
	var b []byte
	copy(b, m.body)
	return NewJSONMessage(b, m.Attempt()+1)
}

func newRabbitMQReceivedMessage(queueName string, body []byte, tag uint64) *RabbitMQReceivedMessage {
	return &RabbitMQReceivedMessage{
		queueName: queueName,
		body:      body,
		tag:       tag,
	}
}
