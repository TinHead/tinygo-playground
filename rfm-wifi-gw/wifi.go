package main

import (
	"errors"
	"github.com/soypat/cyw43439"
	"github.com/soypat/seqs"
	"github.com/soypat/seqs/stacks"
	"log/slog"
	"math/rand"
	"net"
	"net/netip"
	"time"
)

const mtu = cyw43439.MTU
const tcpbufsize = 2030

var (
	myip, _ = netip.ParseAddr("192.168.1.11")
)

func nicLoop(dev *cyw43439.Device, Stack *stacks.PortStack) {
	// Maximum number of packets to queue before sending them.
	const (
		queueSize                = 3
		maxRetriesBeforeDropping = 3
	)
	var queue [queueSize][mtu]byte
	var lenBuf [queueSize]int
	var retries [queueSize]int
	markSent := func(i int) {
		queue[i] = [mtu]byte{} // Not really necessary.
		lenBuf[i] = 0
		retries[i] = 0
	}
	for {
		stallRx := true
		// Poll for incoming packets.
		for i := 0; i < 1; i++ {
			gotPacket, err := dev.PollOne()
			if err != nil {
				println("poll error:", err.Error())
			}
			if !gotPacket {
				break
			}
			stallRx = false
		}

		// Queue packets to be sent.
		for i := range queue {
			if retries[i] != 0 {
				continue // Packet currently queued for retransmission.
			}
			var err error
			buf := queue[i][:]
			lenBuf[i], err = Stack.HandleEth(buf[:])
			if err != nil {
				println("stack error n(should be 0)=", lenBuf[i], "err=", err.Error())
				lenBuf[i] = 0
				continue
			}
			if lenBuf[i] == 0 {
				break
			}
		}
		stallTx := lenBuf == [queueSize]int{}
		if stallTx {
			if stallRx {
				// Avoid busy waiting when both Rx and Tx stall.
				time.Sleep(51 * time.Millisecond)
			}
			continue
		}

		// Send queued packets.
		for i := range queue {
			n := lenBuf[i]
			if n <= 0 {
				continue
			}
			err := dev.SendEth(queue[i][:n])
			if err != nil {
				// Queue packet for retransmission.
				retries[i]++
				if retries[i] > maxRetriesBeforeDropping {
					markSent(i)
					println("dropped outgoing packet:", err.Error())
				}
			} else {
				markSent(i)
			}
		}
	}
}

func ResolveHardwareAddr(stack *stacks.PortStack, ip netip.Addr) ([6]byte, error) {
	if !ip.IsValid() {
		return [6]byte{}, errors.New("invalid ip")
	}
	arpc := stack.ARP()
	arpc.Abort() // Remove any previous ARP requests.
	err := arpc.BeginResolve(ip)
	if err != nil {
		return [6]byte{}, err
	}
	time.Sleep(4 * time.Millisecond)
	// ARP exchanges should be fast, don't wait too long for them.
	const timeout = time.Second
	const maxretries = 20
	retries := maxretries
	for !arpc.IsDone() && retries > 0 {
		retries--
		if retries == 0 {
			return [6]byte{}, errors.New("arp timed out")
		}
		time.Sleep(timeout / maxretries)
	}
	_, hw, err := arpc.ResultAs6()
	return hw, err
}

func initWifi(dev *cyw43439.Device, logger *slog.Logger) *stacks.PortStack {

	// join net
	for {
		// Set ssid/pass in secrets.go
		err := dev.JoinWPA2(ssid, pass)
		if err == nil {
			break
		}
		logger.Error("wifi join failed", slog.String("err", err.Error()))
		time.Sleep(5 * time.Second)
	}
	// wifi connected!
	mac, _ := dev.HardwareAddr6()
	logger.Info("wifi join success!", slog.String("mac", net.HardwareAddr(mac[:]).String()))

	// set ip
	stack := stacks.NewPortStack(stacks.PortStackConfig{
		MAC:             mac,
		MaxOpenPortsUDP: 2,
		MaxOpenPortsTCP: 2,
		MTU:             1420,
		Logger:          logger,
	})
	stack.SetAddr(myip)
	dev.RecvEthHandle(stack.RecvEth)

	// hadle packets
	go nicLoop(dev, stack)

	return stack
}

func setupClient(stack *stacks.PortStack, logger *slog.Logger, broker string) (*stacks.TCPConn, *rand.Rand) {
	start := time.Now()
	rng := rand.New(rand.NewSource(int64(time.Now().Sub(start))))
	clientAddr := netip.AddrPortFrom(stack.Addr(), uint16(rng.Intn(65535-1024)+1024))
	conn, err := stacks.NewTCPConn(stack, stacks.TCPConnConfig{
		TxBufSize: tcpbufsize,
		RxBufSize: tcpbufsize,
	})

	if err != nil {
		panic("conn create:" + err.Error())
	}

	random := rng.Uint32()
	logger.Info("socket:listen")
	svAddr, err := netip.ParseAddrPort(broker)
	if err != nil {
		panic("parsing server address:" + err.Error())
	}
	// Resolver router's hardware address to dial outside our network to internet.
	serverHWAddr, err := ResolveHardwareAddr(stack, svAddr.Addr())
	if err != nil {
		panic("router hwaddr resolving:" + err.Error())
	}
	err = conn.OpenDialTCP(clientAddr.Port(), serverHWAddr, svAddr, seqs.Value(random))
	if err != nil {
		panic("socket dial:" + err.Error())
	}
	retries := 50
	for conn.State() != seqs.StateEstablished && retries > 0 {
		time.Sleep(100 * time.Millisecond)
		retries--
	}
	if retries == 0 {
		logger.Info("socket:no-establish")
		closeConn(conn, "did not establish connection")
	}
	return conn, rng
}

func closeConn(conn *stacks.TCPConn, err string) {
	slog.Error("tcpconn:closing", slog.String("err", err))
	conn.Close()
	for !conn.State().IsClosed() {
		slog.Info("tcpconn:waiting", slog.String("state", conn.State().String()))
		time.Sleep(1000 * time.Millisecond)
	}
}
