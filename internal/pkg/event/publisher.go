package event

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nsqio/go-nsq"
)

type Publisher struct {
	producer *nsq.Producer
	topic    string
}

func NewPublisher(producer *nsq.Producer, topic string) *Publisher {
	return &Publisher{producer: producer, topic: topic}
}

func (p *Publisher) Publish(eventType string, payload map[string]any) {
	if p == nil {
		return
	}
	if payload == nil {
		payload = map[string]any{}
	}
	payload["event_type"] = eventType
	payload["timestamp"] = time.Now().Format(time.RFC3339Nano)

	b, err := json.Marshal(payload)
	if err != nil {
		log.Printf("marshal event failed: %v", err)
		return
	}
	if p.producer == nil || p.topic == "" {
		log.Printf("event(no-nsq): %s", string(b))
		return
	}
	if err = p.producer.Publish(p.topic, b); err != nil {
		log.Printf("publish event failed: %v", err)
	}
}
