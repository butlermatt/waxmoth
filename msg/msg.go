package msg

import (
	"errors"
	"time"
)

const (
	// Various indexes of data
	msgType = iota // Message type
	tType          // Transmission type. MSG type only
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

type MsgType int

const (
	Invalid MsgType = iota // Invalid message type
	Sel                    // Selection Change Message (should never see)
	Id                     // New ID Message (When callsign is set or changed)
	Air                    // New aircraft message (Should never see)
	Sta                    // Status Change Message (Should never see)
	Clk                    // Click message (Should never see)
	Msg                    // Transmission Message (Almost every message seen)
)

type Raw struct {
	origin string
	data   string
}

type Message struct {
	Station   string    // Station reporting the message.
	Type      MsgType   // Type is the MsgType of the message (should always bee Msg)
	TxType    uint8     // TxType is the transmission type of the message. Only for Msg types.
	Icao      uint      // Icao is identification number assigned to the plane by the International Civil Aviation Organization
	GenDate   time.Time // GenDate is the DateTime the message was generated
	LogDate   time.Time // LogDate is the DateTime the message was logged by the Station.
	CallSign  string    // CallSign being broadcast by the aircraft
	Altitude  int       // Altitude of the aircraft (relative to 1013.2mb pressure)
	Speed     float32   // Speed over ground (not indicative of airspeed)
	Track     float32   // Track of aircraft, derived from velocity E/W and velocity N/S
	Latitude  float64   // Latitude position of aircraft.
	Longitude float64   // Longitude position of aircraft.
	Vertical  int       // Vertical rate of climb or decent (64ft resolution?)
	Squawk    uint16    // Squawk assigned to the aircraft and entered by them for each airspace.
	Alert     bool      // Alert indicates that the Squawk has changed.
	Emerg     bool      // Emerg indicates the emergency code has been set.
	SPI       bool      // SPI indicates the Ident transponder has been activated.
	OnGround  bool      // OnGround indicates if the aircraft is reporting being on the ground.
}

// ParseChannel is designed to be run as a goroutine, accepting an inbound and outbound channel for the messages. The
// inbound channel consists of Raw messages received from the station. Outbound consists of parsed Messages.
func ParseChannel(in <-chan *Raw, out chan<- *Message) {

}

// Parse accepts a raw message string and origin, and returns a pointer to a Message, or an error.
func Parse(station string, message string) (*Message, error) {
	return nil, errors.New("not yet implemented")
}

func parseType(s string) MsgType {
	switch s {
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
