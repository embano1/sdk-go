/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	clienttest "github.com/cloudevents/sdk-go/v2/client/test"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"

	protocolamqp "github.com/cloudevents/sdk-go/protocol/amqp/v3"
)

func TestSenderReceiverEvent(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		eventIn = test.ConvertEventExtensionsToString(t, eventIn)
		ctx := context.Background()
		clienttest.SendReceive(t, func() any {
			return protocolFactory(ctx, t)
		}, eventIn, func(e event.Event) {
			test.AssertEventEquals(t, eventIn, test.ConvertEventExtensionsToString(t, e))
		})
	})
}

func protocolFactory(ctx context.Context, t *testing.T) *protocolamqp.Protocol {
	t.Helper()

	const (
		brokerEnvVar = "TEST_AMQP_URI"
		queue        = "amqp-v3-integration-test"
	)

	server := os.Getenv(brokerEnvVar)
	if server == "" {
		t.Errorf("%q must be set", brokerEnvVar)
	}

	p, err := protocolamqp.NewProtocol(ctx, server, queue)
	require.NoError(t, err)

	return p
}
