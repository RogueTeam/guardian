package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/RogueTeam/guardian/sharing"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

var config struct {
	Addr   string
	Target string
}

var (
	addr   multiaddr.Multiaddr
	target multiaddr.Multiaddr
)

func init() {
	set := flag.NewFlagSet("ping", flag.ExitOnError)
	set.StringVar(&config.Addr, "addr", "/ip4/0.0.0.0/tcp/0", "Listen address")
	set.StringVar(&config.Target, "target", "", "Ping target address")
	set.Parse(os.Args[1:])

	var err error
	// Address of the peer
	addr, err = multiaddr.NewMultiaddr(config.Addr)
	if err != nil {
		log.Fatal(err)
	}

	// Ping target
	if config.Target != "" {
		target, err = multiaddr.NewMultiaddr(config.Target)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// /ip4/127.0.0.1/tcp/46337/p2p/12D3KooWBTQQ1BrUuSgd2kHU8sJs1bYoshUcnhvP9nAEA9DoVu5i
func main() {

	l, err := sharing.NewLibP2P(nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Node.Close()

	for _, addr := range l.Addrs {
		fmt.Println("-", addr)
	}

	if config.Target == "" {
		ch := make(chan os.Signal, 1)
		go signal.Notify(ch, syscall.SIGTERM)
		<-ch
		return
	}

	info, err := peer.AddrInfoFromP2pAddr(target)
	if err != nil {
		log.Fatal(err)
	}
	ch, err := l.Ping(100, info)
	if err != nil {
		log.Fatal(err)
	}
	for res := range ch {
		fmt.Println(res.RTT)
	}
}
