package connection

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/networking/codec"
)

// Receive handles an inbound packet.
func (session *Session) Receive(ctx context.Context, packet codec.Packet) error {
	if err := session.markTraffic(EventPacketReceived); err != nil {
		return err
	}

	context := session.context(InboundDirection)
	if context.Disconnected {
		return ErrDisposed
	}

	if session.logger != nil {
		session.logger.Received(context, packet)
	}

	err := session.inbound.Handle(context, packet)
	if errors.Is(err, ErrHandlerNotFound) && session.logger != nil {
		session.logger.Unhandled(context, packet)
	}

	return err
}

// Send handles and writes an outbound packet.
func (session *Session) Send(ctx context.Context, packet codec.Packet) error {
	if err := session.markTraffic(""); err != nil {
		return err
	}

	context := session.context(OutboundDirection)
	if context.Disconnected {
		return ErrDisposed
	}

	if err := session.outbound.Handle(context, packet); err != nil {
		return err
	}

	if err := session.sender(ctx, packet); err != nil {
		return err
	}

	if session.logger != nil {
		session.logger.Sent(context, packet)
	}

	return nil
}
