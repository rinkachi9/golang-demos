package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

// To generate the Go bindings:
// $ go run github.com/cilium/ebpf/cmd/bpf2go -target native Counter bpf/counter.c

func main() {
	// Remove resource limits for kernels < 5.11
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal("Removing memlock:", err)
	}

	// Load pre-compiled implementation (simulated here since we can't run bpf2go freely)
	// In a real workflow, we would use the generated LoadCounterObjects()
	objects := CounterObjects{}
	if err := LoadCounterObjects(&objects, nil); err != nil {
		log.Fatal("Loading eBPF objects:", err)
	}
	defer objects.Close()

	log.Println("Note: This demo requires 'bpf2go' generation step to run.")
	log.Println("Assuming objects loaded, we would attach to an interface.")

	// Example attachment (commented out as it requires valid interface and objects)
	ifaceName := "lo"
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Fatalf("Getting interface %s: %s", ifaceName, err)
	}

	// Attach the program to the interface at the XDP hook
	l, err := link.AttachXDP(link.XDPOptions{
		Program:   objects.CountPackets,
		Interface: iface.Index,
	})
	if err != nil {
		log.Fatal("Attaching XDP:", err)
	}
	defer l.Close()

	log.Printf("Attached XDP program to %s", ifaceName)

	log.Println("Press Ctrl+C to exit...")

	// Periodically read the map
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-stop:
			log.Println("Exiting...")
			return
		case <-ticker.C:
			// Read the map value
			var count uint64
			key := uint32(0)
			if err := objects.PktCount.Lookup(&key, &count); err != nil {
				log.Printf("Map lookup: %s", err)
				continue
			}
			log.Printf("Packet count: %d", count)
		}
	}
}
