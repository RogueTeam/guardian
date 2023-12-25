package sharing_test

import (
	"testing"

	"github.com/RogueTeam/guardian/internal/utils"
	"github.com/RogueTeam/guardian/sharing"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
)

func TestLibP2P_Ping(t *testing.T) {
	t.Run("Succeed", func(t *testing.T) {
		type Test struct {
			Name   string
			ASK    crypto.PrivKey
			ALAddr []multiaddr.Multiaddr
			BSK    crypto.PrivKey
			BLAddr []multiaddr.Multiaddr
		}

		tests := []Test{
			{"Simple", nil, []multiaddr.Multiaddr{
				utils.Must(multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/0"))}, nil,
				[]multiaddr.Multiaddr{utils.Must(multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/0"))},
			},
		}

		for _, test := range tests {
			test := test
			t.Run(test.Name, func(t *testing.T) {
				a, err := sharing.NewLibP2P(test.ASK, test.ALAddr...)
				if err != nil {
					t.Fatalf("expecting no errors: %v", err)
				}
				defer a.Node.Close()

				b, err := sharing.NewLibP2P(test.BSK, test.BLAddr...)
				if err != nil {
					t.Fatalf("expecting no errors: %v", err)
				}
				defer b.Node.Close()

				for _, addr := range b.Addrs {
					pings, err := a.Ping(100, addr)
					if err != nil {
						t.Fatalf("expecting no errors: %v", err)
					}
					for range pings {
					}
				}
			})
		}
	})
}
