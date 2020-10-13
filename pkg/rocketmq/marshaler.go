package rocketmq

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/pkg/errors"
)

// Marshaler marshals Watermill's message to Kafka message.
type Marshaler interface {
	Marshal(topic string, msg *message.Message) ([]*primitive.Message, error)
}

// Unmarshaler unmarshals Kafka's message to Watermill's message.
type Unmarshaler interface {
	Unmarshal([]*primitive.MessageExt) ([]*message.Message, error)
}

// MarshalerUnmarshaler un/marshaler interface for rocketmq message
type MarshalerUnmarshaler interface {
	Marshaler
	Unmarshaler
}

// DefaultMarshaler default message mashaler
type DefaultMarshaler struct{}

// Marshal implement MarshalerUnmarshaler
func (DefaultMarshaler) Marshal(topic string, msg *message.Message) ([]*primitive.Message, error) {
	var rocketmqMsgExts []*primitive.MessageExt
	if len(msg.Payload) > 0 {
		rocketmqMsgExts = primitive.DecodeMessage(msg.Payload)
		if len(rocketmqMsgExts) == 0 {
			return nil, errors.Errorf("empty decoded messages for non-empty msg payload(%s)", msg.UUID)
		}
	}
	var rocketmqMsgs []*primitive.Message
	for _, rocketmqMsgExt := range rocketmqMsgExts {
		rocketmqMsgs = append(rocketmqMsgs, &rocketmqMsgExt.Message)
	}
	if len(rocketmqMsgs) == 0 {
		rocketmqMsg := &primitive.Message{
			Topic: topic,
		}
		if msg.Metadata != nil {
			rocketmqMsg.WithProperties(msg.Metadata)
		}
		rocketmqMsgs = append(rocketmqMsgs, rocketmqMsg)
	}

	return rocketmqMsgs, nil
}

// Unmarshal implement MarshalerUnmarshaler
func (DefaultMarshaler) Unmarshal(msgs []*primitive.MessageExt) ([]*message.Message, error) {
	// TODO
	return nil, nil
}
