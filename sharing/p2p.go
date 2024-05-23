package sharing

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
)

type LibP2P struct {
	Node  host.Host
	Info  peer.AddrInfo
	Addrs []multiaddr.Multiaddr
}

func (l *LibP2P) Ping(n int, info *peer.AddrInfo) (ch <-chan ping.Result, err error) {
	pingService := &ping.PingService{Host: l.Node}
	l.Node.SetStreamHandler(ping.ID, pingService.PingHandler)

	err = l.Node.Connect(context.Background(), *info)
	if err != nil {
		err = fmt.Errorf("failed to connect to peer: %w", err)
		return
	}

	ch = pingService.Ping(context.Background(), info.ID)

	return
}

func NewLibP2P(sk crypto.PrivKey, lAddrs ...multiaddr.Multiaddr) (l *LibP2P, err error) {

	var lAddrsOption libp2p.Option
	if len(lAddrs) > 0 {
		lAddrsOption = libp2p.ListenAddrs(lAddrs...)
	} else {
		lAddrsOption = libp2p.NoListenAddrs
	}

	node, err := libp2p.New(
		libp2p.Identity(sk),
		libp2p.ForceReachabilityPrivate(),
		lAddrsOption,
	)
	if err != nil {
		err = fmt.Errorf("%w: failed to create host: %w", ErrSharingCreation, err)
		return
	}

	idSrv, err := identify.NewIDService(node)
	if err != nil {
		err = fmt.Errorf("%w: failed to prepare id service: %w", ErrSharingCreation, err)
		return
	}
	_, err = holepunch.NewService(node, idSrv)
	if err != nil {
		err = fmt.Errorf("%w: failed to prepare holepunch service: %w", ErrSharingCreation, err)
		return
	}

	l = &LibP2P{
		Node: node,
		Info: peer.AddrInfo{
			ID:    node.ID(),
			Addrs: node.Addrs(),
		},
	}

	l.Addrs, err = peer.AddrInfoToP2pAddrs(&l.Info)
	if err != nil {
		l = nil
		err = fmt.Errorf("%w: failed to prepare addresses: %w", ErrSharingCreation, err)
		return
	}

	return
}
