package msg

import (
	"fmt"
	"time"
)

type Plane struct {
	Stations    []string
	Icao        uint
	CallSign    string
	LastSeen    time.Time
	Locations   []Location
	GroundSpeed float32
	Track       float32
	Altitude    int
	Vertical    int
	Squawk      uint16
	SquawkAlert bool
	Emergency   bool
	IdentAlert  bool
	OnGround    bool
	Messages    []*Message
}

func (p *Plane) AddMessage(m *Message) {
	var contains bool
	for _, s := range p.Stations {
		if s == m.Station {
			contains = true
			break
		}
	}

	if !contains {
		p.Stations = append(p.Stations, m.Station)
	}

	if p.isDuplicate(m) {
		return
	}

	if m.LogDate.After(p.LastSeen) {
		p.LastSeen = m.LogDate
	}

	switch m.TxType {
	case 1:
		p.CallSign = m.CallSign
		fmt.Printf("%x: Callsign: %q\n", p.Icao, p.CallSign)
	case 2:
		p.Altitude = m.Altitude
		p.GroundSpeed = m.Speed
		p.Track = m.Track
		p.Locations = append(p.Locations, m.Location)
		p.OnGround = m.OnGround
		fmt.Printf("%x: Altitude: %d, Speed: %.1f, Track: %.1f, Loc: %f,%f\n", p.Icao, p.Altitude, p.GroundSpeed, p.Track, m.Location.Latitude, m.Location.Longitude)
	case 3:
		p.Altitude = m.Altitude
		p.Locations = append(p.Locations, m.Location)
		p.SquawkAlert = m.Alert
		p.Emergency = m.Emerg
		p.IdentAlert = m.SPI
		p.OnGround = m.OnGround
		fmt.Printf("%x: Altitude: %d, Loc: %f,%f\n", p.Icao, m.Altitude, m.Location.Latitude, m.Location.Longitude)
	case 4:
		p.GroundSpeed = m.Speed
		p.Track = m.Track
		p.Vertical = m.Vertical
		fmt.Printf("%x: Speed: %.1f, Track: %.1f, Vertical: %d\n", p.Icao, m.Speed, m.Track, m.Vertical)
	case 5:
		p.Altitude = m.Altitude
		p.SquawkAlert = m.Alert
		p.IdentAlert = m.SPI
		p.OnGround = m.OnGround
		fmt.Printf("%x: Altitude: %d, SA: %t, IA: %t, Gnd: %t\n", p.Icao, m.Altitude, m.Alert, m.SPI, m.OnGround)
	case 6:
		p.Altitude = m.Altitude
		p.Squawk = m.Squawk
		p.SquawkAlert = m.Alert
		p.Emergency = m.Emerg
		p.IdentAlert = m.SPI
		p.OnGround = m.OnGround
		fmt.Printf("%x: Altitude: %d, Squawk: %d, SA: %t, Emg: %t, IA: %t, Gnd: %t\n", p.Icao, m.Altitude, m.Squawk, m.Alert, m.Emerg, m.SPI, m.OnGround)
	case 7:
		p.Altitude = m.Altitude
		p.OnGround = m.OnGround
		fmt.Printf("%x: Altitude: %d, Gnd: %t\n", p.Icao, m.Altitude, m.OnGround)
	case 8:
		if p.OnGround != m.OnGround {
			p.OnGround = m.OnGround
			fmt.Printf("%x: On Ground: %t\n", p.Icao, m.OnGround)
		}
	}

	p.Messages = append(p.Messages, m)
}

func (p *Plane) isDuplicate(m *Message) bool {
	for i := len(p.Messages) - 1; i >= 0; i-- {
		msg := p.Messages[i]
		if m.Station == msg.Station {
			continue // Skip messages from same station only care about duplicates from other stations.
		}

		// Messages are prior to the existing one shouldn't be a duplicate then
		if msg.GenDate.Before(m.GenDate) {
			return false
		}

		// Skip if the generated dates or Message Type are not the same.
		if !msg.GenDate.Equal(m.GenDate) || msg.TxType != m.TxType {
			continue
		}

		if m.TxType == 1 && m.CallSign == msg.CallSign {
			return true
		}

		if m.TxType == 3 {
			if m.Location.Latitude == msg.Location.Latitude && m.Location.Longitude == msg.Location.Longitude {
				return true
			} // same message type and time but not same location.
			return false
		}

		if m.TxType == 4 {
			if m.Speed == msg.Speed && m.Track == msg.Track && m.Vertical == msg.Vertical {
				return true
			}
			return false
		}

		if (m.TxType == 5 || m.TxType == 7) && m.Altitude == msg.Altitude {
			return true
		}

		if m.TxType == 6 && m.Squawk == msg.Squawk {
			return true
		}

		if m.TxType == 8 && m.OnGround == msg.OnGround {
			return true
		}

		fmt.Printf("Possible duplicates\n%+v\n%+v\n", m, msg)
	}

	return false
}

func (p *Plane) msg2(m *Message) {

}

func New(m *Message) *Plane {
	p := &Plane{
		Icao:      m.Icao,
		Stations:  []string{},
		Locations: []Location{},
		Messages:  []*Message{},
	}

	p.AddMessage(m)

	return p
}

var planes = make(map[uint]*Plane)

func AddMessage(m *Message) {
	p, ok := planes[m.Icao]
	if !ok {
		fmt.Printf("New plane - %x\n", m.Icao)
		planes[m.Icao] = New(m)
	} else {
		p.AddMessage(m)
	}
}
