/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"context"
	"testing"

	"github.com/Azure/go-amqp"

	"github.com/stretchr/testify/require"

	protocolamqp "github.com/cloudevents/sdk-go/protocol/amqp/v3"
	clienttest "github.com/cloudevents/sdk-go/v2/client/test"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
)

func TestSendEvent(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ConvertEventExtensionsToString(t, eventIn)
		clienttest.SendReceive(t, func() interface{} {
			return protocolFactory(t)
		}, eventIn, func(e event.Event) {
			test.AssertEventEquals(t, eventIn, test.ConvertEventExtensionsToString(t, e))
		})
	})
}

func TestSenderReceiverEvent(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ConvertEventExtensionsToString(t, eventIn)
		ctx := context.Background()
		clienttest.SendReceive(t, func() interface{} {
			return protocolFactory(ctx,t)
		}, eventIn, func(e event.Event) {
			test.AssertEventEquals(t, eventIn, test.ConvertEventExtensionsToString(t, e))
		})
	})
}

func protocolFactory(ctx context.Context, t *testing.T) *protocolamqp.Protocol {
	conn, sess, queue := testClient(ctx, t)

	p, err := protocolamqp.NewProtocol(ctx,conn.)
	require.NoError(t, err)

	return p
}

func testClient(ctx context.Context, t *testing.T) (*amqp.Conn, *amqp.Session, string) {
	t.Helper()
	addr := "test-queue"

	s := os.Getenv("TEST_AMQP_URL")
	if u, err := url.Parse(s); err == nil && u.Path != "" {
		addr = u.Path
	}

	conn, err := amqp.Dial(ctx, s, nil)
	if err != nil {
		t.Skipf("ampq.Dial(%#v): %v", s, err)
	}

	session, err := conn.NewSession(ctx, &amqp.SessionOptions{})
	require.NoError(t, err)
	return conn, session, addr
}
