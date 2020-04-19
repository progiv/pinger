package internal

import (
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

type PingConfig struct {
	Host     string
	Count    int
	Proto    int
	Timeout  time.Duration
	Interval time.Duration
	Ttl      int
}

const (
	ProtocolICMP = 1
	// ProtocolIPv6ICMP = 58
	MaxPacketSize = 1024 // Somewhat arbitrary. Should be enough for packages defined by me)
)

type Pinger struct {
	PingConfig
	wg   sync.WaitGroup
	c    *icmp.PacketConn
	addr net.UDPAddr
	stat pingStats
}

func NewPinger(conf PingConfig) *Pinger {
	pinger := Pinger{PingConfig: conf}
	return &pinger
}

func (p *Pinger) Ping() error {
	ips, err := net.LookupIP(p.Host)
	if err != nil {
		return fmt.Errorf("failed to resolve host: %w", err)
	}

	log.Printf("Sending ICMP requests to host: %s, ip: %v\n", p.Host, ips[0])
	p.addr = net.UDPAddr{
		IP: ips[0],
	}

	c, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		return fmt.Errorf("can't create listener: %w", err)
	}
	defer c.Close()
	p.c = c
	err = p.c.IPv4PacketConn().SetTTL(p.Ttl)
	if err != nil {
		log.Fatalf("can't set ttl to %d", p.Ttl)
	}

	defer p.PrintReport()

	done := make(chan interface{})
	cancel := make(chan os.Signal, 1)
	signal.Notify(cancel, os.Interrupt)

	p.wg.Add(1)
	go p.send(done)
	go p.listen(done)

	select {
	case <-cancel:
		log.Println("Received Interrupt signal")
		close(done)
	case <-done:
		// finish
	}
	p.wg.Wait()

	return err
}

func (p *Pinger) send(done <-chan interface{}) {
	defer p.wg.Done()
	for seq := 1; seq <= p.Count || p.Count == 0; seq++ {
		wb, err := echoRequest(seq)
		if err != nil {
			log.Fatalf("create message: %v", err)
		}
		if _, err := p.c.WriteTo(wb, &p.addr); err != nil {
			log.Printf("write: %v", err)
			return
		}
		p.stat.Send()
		select {
		case <-done:
			return
		case <-time.After(p.Interval):
		}
	}
	log.Printf("All %d messages were sent", p.Count)
}

func (p *Pinger) processReply(size int, addr net.Addr, msg *icmp.Message) {
	prefix := fmt.Sprintf("%d bytes from %v, type %d:", size, addr, msg.Type)
	switch m := msg.Body.(type) {
	case *icmp.Echo:
		var pingBody pingMessageBody
		err := pingBody.UnmarshalBinary(m.Data)
		if err != nil {
			log.Println("strange echo body: %w", err)
			return
		}
		elapsed := time.Since(pingBody.time)
		log.Printf("%s icmp_seq=%d, time=%v\n", prefix, m.Seq, elapsed)
		p.stat.Receive(elapsed)
	case *icmp.TimeExceeded:
		log.Printf("%s, TTL", prefix)
		p.stat.Error()
	default:
		log.Printf("%s unsupported ICMP message type", prefix)
		p.stat.Error()
	}
}

func (p *Pinger) listen(done chan<- interface{}) error {
	for seq := 1; seq <= p.Count || p.Count == 0; seq++ {
		buf := make([]byte, MaxPacketSize)
		if err := p.c.SetReadDeadline(time.Now().Add(p.Timeout)); err != nil {
			return fmt.Errorf("can't set read deadline: %w", err)
		}
		size, addr, err := p.c.ReadFrom(buf)

		if err != nil {
			//log.Print(err)
			p.stat.Error()
			netErr, isNetErr := err.(net.Error)
			if isNetErr {
				if netErr.Timeout() {
					log.Printf("%d bytes from %v: Timeout", size, addr)
				}
				if !netErr.Temporary() {
					return fmt.Errorf("listen: %w", netErr)
				}
			} else {
				return fmt.Errorf("listen: %w", err)
			}
		}
		reply, err := icmp.ParseMessage(ProtocolICMP, buf)
		if err != nil {
			p.stat.Error()
			log.Printf("parse message error: %v", err)
		}
		p.processReply(size, addr, reply)
	}
	log.Println("finishing listening")
	close(done)
	return nil
}

func echoRequest(seq int) ([]byte, error) {
	data, err := pingMessageBody{
		time: time.Now(),
	}.BinaryMarshall()
	if err != nil {
		return nil, err
	}

	request := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getegid() & 0xffff,
			Seq:  seq,
			Data: data,
		},
	}
	return request.Marshal(nil)
}

func (p *Pinger) PrintReport() {
	fmt.Printf("--- %s ping statistics ---\n", p.Host)
	fmt.Println(p.stat.Report())
}
