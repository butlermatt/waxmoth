package msg

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	// Various indexes of data
	msgType = iota // Message type
	txType         // Transmission type. MSG type only
	_              // Session Id. Don't care
	_              // Aircraft ID. Don't care (usually 11111)
	icao           // ModeS or ICAO Hex number
	_              // Flight ID. Don't care (usually 11111)
	dGen           // Date message was Generated
	tGen           // Time message was Generated
	dLog           // Date message was logged.
	tLog           // Time message was logged.

	// May not be in every message
	callSign     // Call Sign (Flight ID, Flight Number or Registration)
	alt          // Altitude
	groundSpeed  // Ground Speed (not indicative of air speed)
	track        // Track of aircraft, not heading. Derived from Velocity E/w and Velocity N/S
	latitude     // As it says
	longitude    // As it says
	verticalRate // 64ft resolution
	squawk       // Assigned Mode A squawk code
	squawkAlert  // Flag to indicate squawk change.
	emergency    // Flag to indicate Emergency
	identActive  // Flag to indicate transponder Ident has been activated
	onGround     // Flag to indicate ground squawk switch is active
)

type Type int

const (
	Invalid Type = iota // Invalid message type
	Sel                 // Selection Change Message (should never see)
	Id                  // New ID Message (When callsign is set or changed)
	Air                 // New aircraft message (Should never see)
	Sta                 // Status Change Message (Should never see)
	Clk                 // Click message (Should never see)
	Msg                 // Transmission Message (Almost every message seen)
)

type Raw struct {
	Origin string
	Data   []byte
}

type Location struct {
	Latitude  float64 // Latitude position of the aircraft
	Longitude float64 // Longitude position of the aircraft
}

type Message struct {
	Station  string    // Station reporting the message.
	Type     Type      // Type is the MsgType of the message (should always bee Msg)
	TxType   uint8     // TxType is the transmission type of the message. Only for Msg types.
	Icao     uint      // Icao is identification number assigned to the plane by the International Civil Aviation Organization
	GenDate  time.Time // GenDate is the DateTime the message was generated
	LogDate  time.Time // LogDate is the DateTime the message was logged by the Station.
	CallSign string    // CallSign being broadcast by the aircraft
	Altitude int       // Altitude of the aircraft (relative to 1013.2mb pressure)
	Speed    float32   // Speed over ground (not indicative of airspeed)
	Track    float32   // Track of aircraft, derived from velocity E/W and velocity N/S
	Location Location  // Location is the GPS location of the aircraft.
	Vertical int       // Vertical rate of climb or decent (64ft resolution?)
	Squawk   uint16    // Squawk assigned to the aircraft and entered by them for each airspace.
	Alert    bool      // Alert indicates that the Squawk has changed.
	Emerg    bool      // Emerg indicates the emergency code has been set.
	SPI      bool      // SPI indicates the Ident transponder has been activated.
	OnGround bool      // OnGround indicates if the aircraft is reporting being on the ground.
}

// ParseChannel is designed to be run as a goroutine, accepting an inbound and outbound channel for the messages. The
// inbound channel consists of Raw messages received from the station. Outbound consists of parsed Messages.
func ParseChannel(in <-chan *Raw, out chan<- *Message) {
	for rm := range in {
		m, err := Parse(rm.Origin, rm.Data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse message: %q. error - %v\n", rm.Data, err)
			continue
		}

		out <- m
	}
}

// Parse accepts a raw message string and origin, and returns a pointer to a Message, or an error.
func Parse(station string, message []byte) (*Message, error) {
	msg := &Message{Station: station}

	parts := bytes.Split(message, []byte(","))
	if len(parts) != 22 {
		return nil, fmt.Errorf("message contains incorrect number of parts. expected=22, got=%d", len(parts))
	}

	msg.Type = parseType(parts[msgType])
	if msg.Type == Msg {
		val, err := strconv.ParseUint(string(parts[txType]), 10, 8)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse TxType")
		}
		msg.TxType = uint8(val)
	}

	val, err := strconv.ParseUint(string(parts[icao]), 16, 32)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse ICAO number")
	}
	msg.Icao = uint(val)

	d, err := parseDateTime(parts[dGen : tGen+1])
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse generated date")
	}
	msg.GenDate = d
	d, err = parseDateTime(parts[dLog : tLog+1])
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse logged date")
	}
	msg.LogDate = d

	if msg.Type == Sel || msg.Type == Id {
		msg.CallSign = string(parts[callSign])
	}

	if msg.Type != Msg {
		return msg, nil
	}

	switch msg.TxType {
	case 1:
		msg.CallSign = string(parts[callSign])
	case 2:
		err = parseMsg2(msg, parts)
	case 3:
		err = parseMsg3(msg, parts)
	case 4:
		err = parseMsg4(msg, parts)
	case 5:
		err = parseMsg5(msg, parts)
	case 6:
		err = parseMsg6(msg, parts)
	case 7:
		err = parseMsg7(msg, parts)
	case 8:
		err = parseMsg8(msg, parts)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse msg type: %d", msg.TxType)
	}

	return msg, nil
}

func parseType(s []byte) Type {
	switch string(s) {
	case "SEL":
		return Sel
	case "ID":
		return Id
	case "AIR":
		return Air
	case "STA":
		return Sta
	case "CLK":
		return Clk
	case "MSG":
		return Msg
	}
	return Invalid
}

const dateForm = "2006/01/02 15:04:05.999999999"

func parseDateTime(s [][]byte) (time.Time, error) {
	var d, t string
	if len(s) != 2 {
		return time.Time{}, errors.Errorf("wrong size arguments. slice len expected=2 got=%d", len(s))
	}

	d = string(s[0])
	if len(s[1]) > 18 {
		// SBS-1 BaseStation format will sometimes add more nanosecond precision
		// than go can handle so trim it if needed.
		t = string(s[1][0:18])
	} else {
		t = string(s[1])
	}

	date, err := time.Parse(dateForm, d+" "+t)
	if err != nil {
		return time.Time{}, err
	}
	return date, nil
}

func parseMsg2(m *Message, parts [][]byte) error {
	alt, err := strconv.ParseInt(string(parts[alt]), 10, strconv.IntSize)
	if err != nil {
		return errors.Wrap(err, "failed to parse altitude")
	}
	m.Altitude = int(alt)

	f, err := strconv.ParseFloat(string(parts[groundSpeed]), 32)
	if err != nil {
		return errors.Wrap(err, "failed to parse ground speed")
	}
	m.Speed = float32(f)

	f, err = strconv.ParseFloat(string(parts[track]), 32)
	if err != nil {
		return errors.Wrap(err, "failed to parse track")
	}
	m.Track = float32(f)

	lat, err := strconv.ParseFloat(string(parts[latitude]), 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse latitude")
	}
	lon, err := strconv.ParseFloat(string(parts[longitude]), 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse longitude")
	}
	m.Location = Location{Latitude: lat, Longitude: lon}

	m.OnGround = string(parts[onGround]) == "1"

	return nil
}

func parseMsg3(m *Message, parts [][]byte) error {
	alt, err := strconv.ParseInt(string(parts[alt]), 10, strconv.IntSize)
	if err != nil {
		return errors.Wrap(err, "failed to parse altitude")
	}
	m.Altitude = int(alt)

	lat, err := strconv.ParseFloat(string(parts[latitude]), 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse latitude")
	}

	lon, err := strconv.ParseFloat(string(parts[longitude]), 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse longitude")
	}
	m.Location = Location{Latitude: lat, Longitude: lon}

	m.Alert = string(parts[squawkAlert]) == "1"
	m.Emerg = string(parts[emergency]) == "1"
	m.SPI = string(parts[identActive]) == "1"
	m.OnGround = string(parts[onGround]) == "1"

	return nil
}

func parseMsg4(m *Message, parts [][]byte) error {
	if len(parts[groundSpeed]) > 0 {
		f, err := strconv.ParseFloat(string(parts[groundSpeed]), 32)
		if err != nil {
			return errors.Wrap(err, "failed to parse ground speed")
		}
		m.Speed = float32(f)
	}

	if len(parts[track]) > 0 {
		f, err := strconv.ParseFloat(string(parts[track]), 32)
		if err != nil {
			return errors.Wrap(err, "failed to parse track")
		}
		m.Track = float32(f)
	}

	if len(parts[verticalRate]) > 0 {
		i, err := strconv.ParseInt(string(parts[verticalRate]), 10, strconv.IntSize)
		if err != nil {
			return errors.Wrap(err, "failed to parse vertical rate")
		}
		m.Vertical = int(i)
	}

	return nil
}

func parseMsg5(m *Message, parts [][]byte) error {
	alt, err := strconv.ParseInt(string(parts[alt]), 10, strconv.IntSize)
	if err != nil {
		return errors.Wrap(err, "failed to parse altitude")
	}
	m.Altitude = int(alt)

	m.Alert = string(parts[squawkAlert]) == "1"
	m.SPI = string(parts[identActive]) == "1"
	m.OnGround = string(parts[onGround]) == "1"

	return nil
}

func parseMsg6(m *Message, parts [][]byte) error {
	if len(parts[alt]) > 0 {
		alt, err := strconv.ParseInt(string(parts[alt]), 10, strconv.IntSize)
		if err != nil {
			return errors.Wrap(err, "failed to parse altitude")
		}
		m.Altitude = int(alt)
	}

	i, err := strconv.ParseUint(string(parts[squawk]), 10, 16)
	if err != nil {
		return errors.Wrap(err, "failed to parse squawk")
	}
	m.Squawk = uint16(i)

	m.Alert = string(parts[squawkAlert]) == "1"
	m.Emerg = string(parts[emergency]) == "1"
	m.SPI = string(parts[identActive]) == "1"
	m.OnGround = string(parts[onGround]) == "1"

	return nil
}

func parseMsg7(m *Message, parts [][]byte) error {
	alt, err := strconv.ParseInt(string(parts[alt]), 10, strconv.IntSize)
	if err != nil {
		return errors.Wrap(err, "failed to parse altitude")
	}
	m.Altitude = int(alt)

	m.OnGround = string(parts[onGround]) == "1"
	return nil
}

func parseMsg8(m *Message, parts [][]byte) error {
	m.OnGround = string(parts[onGround]) == "1"
	return nil
}
