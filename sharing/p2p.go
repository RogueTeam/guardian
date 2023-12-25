package sharing

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
)

type LibP2P struct {
	Node  host.Host
	Info  peer.AddrInfo
	Addrs []multiaddr.Multiaddr
}

func (l *LibP2P) Ping(n int, peerAddr multiaddr.Multiaddr) (pings chan time.Duration, err error) {
	pingService := &ping.PingService{Host: l.Node}
	l.Node.SetStreamHandler(ping.ID, pingService.PingHandler)

	peer, err := peer.AddrInfoFromP2pAddr(peerAddr)
	if err != nil {
		panic(err)
	}
	if err := l.Node.Connect(context.Background(), *peer); err != nil {
		panic(err)
	}

	pings = make(chan time.Duration, 10)
	go func() {
		defer close(pings)

		ch := pingService.Ping(context.Background(), peer.ID)
		for i := 0; i < n; i++ {
			res := <-ch
			pings <- res.RTT
			if res.Error != nil {
				err = fmt.Errorf("failed to ping: %w", err)
				break
			}
		}
	}()

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
		lAddrsOption,
	)
	if err != nil {
		err = fmt.Errorf("%w: failed to create host: %w", ErrSharingCreation, err)
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
