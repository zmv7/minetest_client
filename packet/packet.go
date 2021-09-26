package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type Packet struct {
	Command      Command
	PacketType   PacketType
	SubType      PacketType
	ControlType  ControlType
	PeerID       uint16
	SeqNr        uint16
	Channel      uint8
	Payload      []byte
	CommandID    uint16
	SplitPayload *SplitPayload
}

type SplitPayload struct {
	SeqNr       uint16
	ChunkCount  uint16
	ChunkNumber uint16
	Data        []byte
}

func (spl *SplitPayload) String() string {
	return fmt.Sprintf("{SplitPayload SeqNr=%d, Chunk=%d/%d, #Data=%d}",
		spl.SeqNr, spl.ChunkNumber+1, spl.ChunkCount, len(spl.Data))
}

func Parse(data []byte) (*Packet, error) {
	p := &Packet{}
	err := p.UnmarshalPacket(data)
	return p, err
}

func CreateReliable(peerId uint16, seqNr uint16, command Command) *Packet {
	return &Packet{
		PacketType: Reliable,
		SubType:    Original,
		Command:    command,
		PeerID:     peerId,
		SeqNr:      seqNr,
		Channel:    1,
	}
}

func CreateOriginal(peerId uint16, seqNr uint16, command Command) *Packet {
	return &Packet{
		PacketType: Original,
		Command:    command,
		PeerID:     peerId,
		SeqNr:      seqNr,
		Channel:    1,
	}
}

func CreateControl(peerId uint16, seqNr uint16, controlType ControlType) *Packet {
	return &Packet{
		PacketType:  Control,
		ControlType: controlType,
		PeerID:      peerId,
		SeqNr:       seqNr,
	}
}

func CreatePacket(packetType PacketType, subType PacketType, peerId uint16, seqNr uint16, command Command) *Packet {
	return &Packet{
		PacketType: packetType,
		SubType:    subType,
		Command:    command,
		PeerID:     peerId,
		SeqNr:      seqNr,
	}
}

func (p *Packet) MarshalPacket() ([]byte, error) {
	packet := make([]byte, 4+2+1+1)
	copy(packet[0:], ProtocolID)
	binary.BigEndian.PutUint16(packet[4:], p.PeerID)
	packet[6] = p.Channel
	packet[7] = byte(p.PacketType)

	if p.PacketType == Reliable {
		bytes := make([]byte, 5)
		binary.BigEndian.PutUint16(bytes, p.SeqNr)
		bytes[2] = byte(p.SubType)
		binary.BigEndian.PutUint16(bytes[3:], p.Command.GetCommandId())

		packet = append(packet, bytes...)
		payload, err := p.Command.MarshalPacket()
		if err != nil {
			return nil, err
		}
		packet = append(packet, payload...)

	} else if p.PacketType == Control {
		bytes := make([]byte, 3)
		bytes[0] = byte(p.ControlType)
		binary.BigEndian.PutUint16(bytes[1:], p.SeqNr)
		packet = append(packet, bytes...)

	} else if p.PacketType == Original {
		bytes := make([]byte, 2)
		binary.BigEndian.PutUint16(bytes, p.Command.GetCommandId())
		packet = append(packet, bytes...)
		payload, err := p.Command.MarshalPacket()
		if err != nil {
			return nil, err
		}
		packet = append(packet, payload...)

	}

	return packet, nil
}

func (p *Packet) UnmarshalPacket(data []byte) error {
	if len(data) < 5 {
		return errors.New("invalid packet length")
	}

	for i, sig := range ProtocolID {
		if data[i] != sig {
			return errors.New("invalid protocol_id")
		}
	}

	p.PeerID = binary.BigEndian.Uint16(data[4:])
	p.Channel = data[6]
	p.PacketType = PacketType(data[7])

	if p.PacketType == Reliable {
		p.SeqNr = binary.BigEndian.Uint16(data[8:])
		p.SubType = PacketType(data[10])

		switch p.SubType {
		case Control:
			p.ControlType = ControlType(data[11])

			if p.ControlType == SetPeerID {
				p.PeerID = binary.BigEndian.Uint16(data[12:])
			}
		case Split:
			spl := &SplitPayload{}
			spl.SeqNr = binary.BigEndian.Uint16(data[11:])
			spl.ChunkCount = binary.BigEndian.Uint16(data[13:])
			spl.ChunkNumber = binary.BigEndian.Uint16(data[15:])
			spl.Data = data[17:]

			p.SplitPayload = spl
		default:
			fmt.Printf("Unknown packet: %s\n", fmt.Sprint(data))
			//TODO: split
			p.CommandID = binary.BigEndian.Uint16(data[11:])
			p.Payload = data[13:]
			cmd, err := CreateCommand(p.CommandID, p.Payload)
			p.Command = cmd
			return err
		}
	}

	return nil
}

func (p *Packet) String() string {
	return fmt.Sprintf("{Packet Type: %s, PeerID: %d, Channel: %d, SeqNr: %d,"+
		"Subtype: %s, ControlType: %s CommandID: %d, Command: %s, SplitPayload: %s}",
		p.PacketType, p.PeerID, p.Channel, p.SeqNr,
		p.SubType, p.ControlType, p.CommandID, p.Command, p.SplitPayload)
}
