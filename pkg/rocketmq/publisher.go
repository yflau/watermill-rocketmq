package rocketmq

import (
	"context"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/pkg/errors"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

// Publisher the rocketmq publisher
type Publisher struct {
	config   PublisherConfig
	producer rocketmq.Producer
	logger   watermill.LoggerAdapter

	closed bool
}

// NewPublisher creates a new RocketMQ Publisher.
func NewPublisher(
	config PublisherConfig,
	logger watermill.LoggerAdapter,
) (*Publisher, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = watermill.NopLogger{}
	}
	producer, err := rocketmq.NewProducer(config.Options()...)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create RocketMQ producer")
	}
	return &Publisher{
		config:   config,
		producer: producer,
		logger:   logger,
	}, nil
}

// PublisherConfig the rocketmq publisher config
type PublisherConfig struct {
	GroupName             string
	InstanceName          string
	Namespace             string
	SendMsgTimeout        time.Duration
	VIPChannelEnabled     bool
	RetryTimes            int
	Interceptors          []primitive.Interceptor
	Selector              producer.QueueSelector
	Credentials           *primitive.Credentials
	DefaultTopicQueueNums int
	CreateTopicKey        string
	// NsResolver            primitive.NsResolver
	// NameServer            primitive.NamesrvAddr
	// NameServerDomain      string

	// Marshaler is used to marshal messages from Watermill format into Kafka format.
	Marshaler Marshaler
}

// Options generate options
func (c *PublisherConfig) Options() []producer.Option {
	var opts []producer.Option
	if c.GroupName != "" {
		opts = append(opts, producer.WithGroupName(c.GroupName))
	}
	if c.InstanceName != "" {
		opts = append(opts, producer.WithInstanceName(c.InstanceName))
	}
	if c.Namespace != "" {
		opts = append(opts, producer.WithNamespace(c.Namespace))
	}
	if c.SendMsgTimeout > 0 {
		opts = append(opts, producer.WithSendMsgTimeout(c.SendMsgTimeout))
	}
	if c.VIPChannelEnabled {
		opts = append(opts, producer.WithVIPChannel(c.VIPChannelEnabled))
	}
	if c.RetryTimes > 0 {
		opts = append(opts, producer.WithRetry(c.RetryTimes))
	}
	if len(c.Interceptors) > 0 {
		opts = append(opts, producer.WithInterceptor(c.Interceptors...))
	}
	if c.Selector != nil {
		opts = append(opts, producer.WithQueueSelector(c.Selector))
	}
	if c.Credentials != nil {
		opts = append(opts, producer.WithCredentials(*c.Credentials))
	}
	if c.DefaultTopicQueueNums > 0 {
		opts = append(opts, producer.WithDefaultTopicQueueNums(c.DefaultTopicQueueNums))
	}
	if c.CreateTopicKey != "" {
		opts = append(opts, producer.WithCreateTopicKey(c.CreateTopicKey))
	}
	return nil
}

// Validate validate publisher config
func (c PublisherConfig) Validate() error {
	return nil
}

// Publish publishes message to Kafka.
//
// Publish is blocking and wait for ack from Kafka.
// When one of messages delivery fails - function is interrupted.
func (p *Publisher) Publish(topic string, msgs ...*message.Message) error {
	if p.closed {
		return errors.New("publisher closed")
	}
	logFields := make(watermill.LogFields, 4)
	logFields["topic"] = topic
	for _, msg := range msgs {
		logFields["message_uuid"] = msg.UUID
		p.logger.Trace("Sending message to RocketMQ", logFields)
		rocketmqMsgs, err := p.config.Marshaler.Marshal(topic, msg)
		if err != nil {
			return errors.Wrapf(err, "cannot marshal message %s", msg.UUID)
		}
		for _, rocketmqMsg := range rocketmqMsgs {
			result, err := p.producer.SendSync(context.Background(), rocketmqMsg)
			if err != nil {
				return errors.WithMessagef(err, "send sync msg %s failed", msg.UUID)
			}
			logFields["send_status"] = result.Status
			logFields["msg_id"] = result.MsgID
			logFields["offset_msg_id"] = result.OffsetMsgID
			logFields["queue_offset"] = result.QueueOffset
			logFields["message_queue"] = result.MessageQueue.String()
			logFields["transaction_id"] = result.TransactionID
			logFields["region_id"] = result.RegionID
			logFields["trace_on"] = result.TraceOn
			p.logger.Trace("Message sent to RocketMQ", logFields)
		}
	}

	return nil
}

// Close closes the publisher
func (p *Publisher) Close() error {
	if p.closed {
		return nil
	}
	p.closed = true

	if err := p.producer.Shutdown(); err != nil {
		return errors.Wrap(err, "cannot close Kafka producer")
	}

	return nil
}
