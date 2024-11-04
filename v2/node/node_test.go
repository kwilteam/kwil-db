package node

import (
	"context"
	"crypto/rand"
	"fmt"
	"kwil/log"
	"kwil/node/types"
	"net"
	"os"
	"os/signal"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	mock "github.com/libp2p/go-libp2p/p2p/net/mock"
	ma "github.com/multiformats/go-multiaddr"
)

var blackholeIP6 = net.ParseIP("100::")

func newTestHost(t *testing.T, mn mock.Mocknet) (host.Host, error) {
	privKey, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	id, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		t.Fatalf("Failed to get private key: %v", err)
	}
	suffix := id
	if len(id) > 8 {
		suffix = id[len(id)-8:]
	}
	ip := append(net.IP{}, blackholeIP6...)
	copy(ip[net.IPv6len-len(suffix):], suffix)
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip6/%s/tcp/4242", ip))
	if err != nil {
		t.Fatalf("Failed to create multiaddress: %v", err)
	}
	// t.Log(addr) // e.g. /ip6/100::1bb1:760e:df55:9ed1/tcp/4242
	return mn.AddPeer(privKey, addr)
}

func TestDualNodeMocknet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	mn := mock.New()

	h1, err := newTestHost(t, mn)
	if err != nil {
		t.Fatalf("Failed to add peer to mocknet: %v", err)
	}

	h2, err := newTestHost(t, mn)
	if err != nil {
		t.Fatalf("Failed to add peer to mocknet: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cancel()
	}()

	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})

	log1 := log.New(log.WithName("NODE1"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
	node1, err := NewNode(".testnet", WithLogger(log1), WithHost(h1), WithRole(types.RoleLeader))
	if err != nil {
		t.Fatalf("Failed to create Node 1: %v", err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer os.RemoveAll(node1.Dir())
		node1.Start(ctx)
	}()

	log2 := log.New(log.WithName("NODE2"), log.WithWriter(os.Stdout), log.WithLevel(log.LevelDebug), log.WithFormat(log.FormatUnstructured))
	node2, err := NewNode(".n2", WithLogger(log2), WithHost(h2), WithRole(types.RoleValidator))
	if err != nil {
		t.Fatalf("Failed to create Node 2: %v", err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer os.RemoveAll(node2.Dir())
		node2.Start(ctx)
	}()

	// Link and connect the hosts
	if err := mn.LinkAll(); err != nil {
		t.Fatalf("Failed to link hosts: %v", err)
	}
	if err := mn.ConnectAllButSelf(); err != nil {
		t.Fatalf("Failed to connect hosts: %v", err)
	}

	// n1 := mn.Net(h1.ID())
	// links := mn.LinksBetweenPeers(h1.ID(), h2.ID())
	// ln := links[0]
	// net := ln.Networks()[0]
	// peers := net.Peers()
	// t.Log(peers)

	// run for a bit, checks stuff, do tests, like ensure blocks mine (TODO)...
	time.Sleep(3 * time.Second)
	cancel()
	wg.Wait()

}

/*func TestStreamFake(t *testing.T) {
	// Open a stream from h2 to h1
	stream2to1, err := h2.NewStream(ctx, h1.ID(), ProtocolIDBlock)
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}
	t.Cleanup(func() {
		stream2to1.Close()
	})

	stream1to2 := mocknet.StreamComplement(stream2to1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		stream2to1.SetWriteDeadline(time.Now().Add(time.Second)) // no op with mock stream!

		// Write data to the stream and check the response
		n, err := stream2to1.Write([]byte("Hello from h2"))
		if err != nil {
			t.Errorf("Failed to write to stream: %v", err)
		}
		t.Log("Write", n)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		stream1to2.SetReadDeadline(time.Now().Add(time.Second)) // no op with mock stream!

		// Read the response from the stream
		buf := make([]byte, 6)
		n, err := stream1to2.Read(buf)
		if err != nil {
			t.Errorf("Failed to read from stream: %v", err)
		}
		t.Log("Read", n)

		// Verify the response
		expectedOutput := "Received: Hello from h2"
		if output := string(buf[:n]); output != expectedOutput {
			t.Errorf("unexpected output: got %q, want %q", output, expectedOutput)
		}
		cancel()
	}()

	wg.Wait()
}*/
