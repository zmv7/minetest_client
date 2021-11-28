package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"minetest_client/commands"
	"minetest_client/packet"
	"net"
	"sync"
)

type Client struct {
	conn          net.Conn
	Host          string
	Port          int
	PeerID        uint16
	sph           *packet.SplitpacketHandler
	netrx         chan []byte
	listeners     []chan interface{}
	listener_lock *sync.RWMutex
}

func NewClient(host string, port int) *Client {
	return &Client{
		Host:          host,
		Port:          port,
		sph:           packet.NewSplitPacketHandler(),
		netrx:         make(chan []byte, 1000),
		listeners:     make([]chan interface{}, 0),
		listener_lock: &sync.RWMutex{},
	}
}

func (c *Client) Start() error {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return err
	}
	c.conn = conn
	go c.rxLoop()
	go c.parseLoop()

	return nil
}

func (c *Client) Stop() error {
	err := c.Send(packet.CreateControl(c.PeerID, packet.Disco))
	if err != nil {
		return err
	}
	close(c.netrx)
	return c.conn.Close()
}

func (c *Client) Init() error {
	peerInit := packet.CreateReliable(0, []byte{0, 0})
	peerInit.Channel = 0
	return c.Send(peerInit)
}

func (c *Client) AddListener(ch chan interface{}) {
	c.listener_lock.Lock()
	defer c.listener_lock.Unlock()
	c.listeners = append(c.listeners, ch)
}

func (c *Client) emitCommand(cmd interface{}) {
	c.listener_lock.RLock()
	defer c.listener_lock.RUnlock()

	for _, ch := range c.listeners {
		select {
		case ch <- cmd:
		default:
		}
	}
}

func (c *Client) SendOriginalCommand(cmd packet.Command) error {
	//fmt.Printf("Sending original command: %s\n", cmd)

	payload, err := packet.CreatePayload(cmd)
	if err != nil {
		return err
	}

	pkg := packet.CreateOriginal(c.PeerID, payload)
	return c.Send(pkg)
}

func (c *Client) SendCommand(cmd packet.Command) error {
	//fmt.Printf("Sending command: %s\n", cmd)

	payload, err := packet.CreatePayload(cmd)
	if err != nil {
		return err
	}

	if len(payload) < packet.MaxPacketLength {
		// one packet
		pkg := packet.CreateReliable(c.PeerID, payload)
		return c.Send(pkg)

	} else {
		// split packet
		pkgs, err := c.sph.SplitPayload(payload)
		if err != nil {
			return err
		}

		for _, pkg := range pkgs {
			pkg.PeerID = c.PeerID
			pkg.Channel = 1
			pkg.SeqNr = packet.NextSequenceNr()

			err = c.Send(pkg)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (c *Client) Send(packet *packet.Packet) error {
	data, err := packet.MarshalPacket()
	if err != nil {
		return err
	}
	//fmt.Printf("Sending packet: %s\n", packet)
	//fmt.Printf("Sending raw: %s\n", fmt.Sprint(data))

	_, err = c.conn.Write(data)
	return err
}

func (c *Client) handleCommandPayload(payload []byte) error {
	commandId := binary.BigEndian.Uint16(payload[0:])
	commandPayload := payload[2:]
	var err error

	//fmt.Printf("Received commandId: %d\n", commandId)

	switch commandId {
	case commands.ServerCommandSetPeer:
		cmd := &commands.ServerSetPeer{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	case commands.ServerCommandHello:
		cmd := &commands.ServerHello{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			packet.ResetSeqNr(65500)
			c.emitCommand(cmd)
		}

	case commands.ServerCommandSRPBytesSB:
		cmd := &commands.ServerSRPBytesSB{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	case commands.ServerCommandAuthAccept:
		cmd := &commands.ServerAuthAccept{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	case commands.ServerCommandAnnounceMedia:
		cmd := &commands.ServerAnnounceMedia{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	case commands.ServerCommandCSMRestrictionFlags:
		cmd := &commands.ServerCSMRestrictionFlags{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	case commands.ServerCommandBlockData:
		cmd := &commands.ServerBlockData{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	case commands.ServerCommandTimeOfDay:
		cmd := &commands.ServerTimeOfDay{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	case commands.ServerCommandChatMessage:
		cmd := &commands.ServerChatMessage{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	case commands.ServerCommandAddParticleSpawner:
		cmd := &commands.ServerAddParticleSpawner{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	case commands.ServerCommandDetachedInventory:
		cmd := &commands.ServerDetachedInventory{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	case commands.ServerCommandHudChange:
		cmd := &commands.ServerHudChange{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	case commands.ServerCommandActiveObjectMessage:
		cmd := &commands.ServerActiveObjectMessage{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	case commands.ServerCommandDeleteParticleSpawner:
		cmd := &commands.ServerDeleteParticleSpawner{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	case commands.ServerCommandMovePlayer:
		cmd := &commands.ServerMovePlayer{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	case commands.ServerCommandMedia:
		cmd := &commands.ServerMedia{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	case commands.ServerCommandAccessDenied:
		cmd := &commands.ServerAccessDenied{}
		if err = cmd.UnmarshalPacket(commandPayload); err == nil {
			c.emitCommand(cmd)
		}

	default:
		fmt.Printf("Unknown command received: %d\n", commandId)
	}

	return err
}

func (c *Client) onReceive(p *packet.Packet) error {
	//fmt.Printf("Received packet: %s\n", p)

	if p.PacketType == packet.Reliable || p.PacketType == packet.Original {
		if p.ControlType == packet.SetPeerID {
			c.PeerID = p.PeerID
			cmd := &commands.ServerSetPeer{
				PeerID: p.PeerID,
			}

			c.emitCommand(cmd)
		}
	}

	// send ack
	if p.PacketType == packet.Reliable {
		ack := packet.CreateControlAck(c.PeerID, p)
		ack.Channel = p.Channel
		if err := c.Send(ack); err != nil {
			return err
		}
	}

	if p.SubType == packet.Reliable || p.SubType == packet.Original {
		if err := c.handleCommandPayload(p.Payload); err != nil {
			return err
		}
	}

	if p.SubType == packet.Split {
		//shove into list
		data := c.sph.AddPacket(p.SplitPayload)

		if data != nil {
			if err := c.handleCommandPayload(data); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) parseLoop() {
	for buf := range c.netrx {
		//fmt.Printf("Received raw: %s\n", fmt.Sprint(buf[:len]))

		p, err := packet.Parse(buf)
		if err != nil {
			panic(err)
		}

		err = c.onReceive(p)
		if err != nil {
			panic(err)
		}
	}
}

func (c *Client) rxLoop() {
	for {
		buf := make([]byte, 1024)
		len, err := c.conn.Read(buf)
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			panic(err)
		}

		c.netrx <- buf[:len]
	}
}
