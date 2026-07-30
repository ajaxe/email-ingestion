package redis

type StreamData struct {
	MessageID  string
	Data       any
	StreamName string
	GroupName  string
	ConsumerID string
}
