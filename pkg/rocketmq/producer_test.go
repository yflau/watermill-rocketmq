package rocketmq_test

import (
	"testing"

	"github.com/yflau/watermill-rocketmq/pkg/rocketmq"
)

func TestNewProducer(t *testing.T) {
	pub := rocketmq.NewPublisher(rocketmq.ProducerConfig{}, nil)
	if pub.SendMode != "sync" {
		t.Fatal("should ==")
	}
	if pub.SendAsyncCallback == nil {
		t.Fatal("should not nil")
	}
}
