package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	mrand2 "math/rand/v2"
	"os"
	"os/signal"
	"syscall"

	"p2p/node"
	dummyce "p2p/node/consensus/mock"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("shutdown signal received")
		cancel()
	}()

	if err := run(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

var (
	key       string
	port      uint64
	connectTo string
	noPex     bool
	leader    bool
	numVals   int
)

func run(ctx context.Context) error {
	flag.StringVar(&key, "key", "", "private key bytes (hexadecimal), empty is pseudo-random")
	flag.Uint64Var(&port, "port", 0, "listen port (0 for random)")
	flag.StringVar(&connectTo, "connect", "", "peer multiaddr to connect to")
	flag.BoolVar(&noPex, "no-pd", false, "disable peer discover")
	flag.BoolVar(&leader, "leader", false, "make this node produce blocks (should only be one in a network)")
	flag.IntVar(&numVals, "v", 1, "number of validators (all peers are validators!)")
	flag.Parse()

	dummyce.NumValidatorsFake = numVals

	rr := rand.Reader
	if port != 0 { // deterministic key based on port for testing
		// rr = mrand.New(mrand.NewSource(int64(port)))
		var seed [32]byte
		binary.LittleEndian.PutUint64(seed[:], port)
		seed = sha256.Sum256(seed[:])
		log.Printf("seed: %x", seed)
		rr = mrand2.NewChaCha8(seed)
		// var buf bytes.Buffer
		// buf.Write(seed[:])
		// buf.Write(seed[:])
		// rr = &buf
	}

	var rawKey []byte
	if key == "" {
		privKey := node.NewKey(rr)
		rawKey, _ = privKey.Raw()
		log.Printf("priv key: %x", rawKey)
	} else {
		var err error
		rawKey, err = hex.DecodeString(key)
		if err != nil {
			return err
		}
	}

	node, err := node.NewNode(port, rawKey, leader, !noPex)
	if err != nil {
		return err
	}

	addr := node.Addr()
	log.Printf("to connect: %s", addr)

	var bootPeers []string
	if connectTo != "" {
		bootPeers = append(bootPeers, connectTo)
	}
	if err = node.Start(ctx, bootPeers...); err != nil {
		return err
	}
	// Start is blocking, for now.

	return nil
}
