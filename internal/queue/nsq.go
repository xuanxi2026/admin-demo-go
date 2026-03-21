package queue

import (
	"encoding/json"

	"github.com/nsqio/go-nsq"
)

func PublishJSON(producer *nsq.Producer, topic string, payload any) error {
	if producer == nil || topic == "" {
		return nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return producer.Publish(topic, b)
}
