package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gofr.dev/pkg/datastore/pubsub"
	"gofr.dev/pkg/errors"
	"gofr.dev/pkg/gofr"
	"gofr.dev/pkg/gofr/request"
	"gofr.dev/pkg/gofr/types"

	"github.com/stretchr/testify/assert"
)

type mockPubSub struct {
	id string
}

func (m *mockPubSub) PublishEventWithOptions(string, interface{}, map[string]string, *pubsub.PublishOptions) error {
	if m.id == "1" {
		return errors.EntityNotFound{ID: "1"}
	}

	return nil
}

func (m *mockPubSub) PublishEvent(string, interface{}, map[string]string) error {
	if m.id == "1" {
		return errors.EntityNotFound{ID: "1"}
	}

	return nil
}

func (m *mockPubSub) Subscribe() (*pubsub.Message, error) {
	return &pubsub.Message{}, nil
}

func (m *mockPubSub) SubscribeWithCommit(pubsub.CommitFunc) (*pubsub.Message, error) {
	return &pubsub.Message{}, nil
}

func (m *mockPubSub) Bind([]byte, interface{}) error {
	return nil
}

func (m *mockPubSub) Ping() error {
	return nil
}

func (m *mockPubSub) HealthCheck() types.Health {
	return types.Health{}
}

func (m *mockPubSub) IsSet() bool {
	return true
}

func (m *mockPubSub) CommitOffset(pubsub.TopicPartition) {
}

func TestProducerHandler(t *testing.T) {
	app := gofr.New()
	m := mockPubSub{}
	app.PubSub = &m

	tests := []struct {
		desc string
		id   string
		resp interface{}
		err  error
	}{
		{"error from publisher", "1", nil, errors.EntityNotFound{ID: "1"}},
		{"success", "123", nil, nil},
	}

	req := httptest.NewRequest(http.MethodGet, "http://dummy", nil)
	context := gofr.NewContext(nil, request.NewHTTPRequest(req), app)

	for i, tc := range tests {
		context.SetPathParams(map[string]string{
			"id": tc.id,
		})

		m.id = tc.id

		resp, err := Producer(context)

		assert.Equal(t, tc.err, err, "TEST[%d], failed.\n%s", i, tc.desc)

		assert.Equal(t, tc.resp, resp, "TEST[%d], failed.\n%s", i, tc.desc)
	}
}

func TestConsumerHandler(t *testing.T) {
	app := gofr.New()

	app.PubSub = &mockPubSub{}

	ctx := gofr.NewContext(nil, nil, app)

	_, err := Consumer(ctx)
	if err != nil {
		t.Errorf("TEST, Failed.\tExpected %v\tGot %v\n", nil, err)
	}
}
