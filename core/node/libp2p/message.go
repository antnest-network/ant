package libp2p

import (
	"context"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/antnest-network/ant-proto"
	"go.uber.org/fx"
)

func Messenger(lc fx.Lifecycle, peerHost host.Host) proto.Messenger {
	ctx, cancel := context.WithCancel(context.Background())
	messenger := proto.NewAntMessenger(ctx, peerHost, []protocol.ID{
		proto.ProtocolPingMessage,
		proto.ProtocolPongMessage,
		proto.ProtocolPushBlockMessage,
		proto.ProtocolMigrateBlockMessage,
		proto.ProtocolCheque,
		proto.ProtocolQueens,
	})

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			cancel()
			return nil
		},
	})
	return messenger
}
