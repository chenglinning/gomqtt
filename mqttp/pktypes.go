package mqttp

const PRONAME string = "MQTT"

const (
	TBD		byte = 0
	MQTT311 byte = 4
	MQTT50  byte = 5
)

const (
	RESERVED byte = iota
	CONNECT
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
	AUTH
)

// PacketNames maps the constants for each of the MQTT packet types
// to a string representation of their name.
var PacketNames = map[byte]string{
	1:  "CONNECT",
	2:  "CONNACK",
	3:  "PUBLISH",
	4:  "PUBACK",
	5:  "PUBREC",
	6:  "PUBREL",
	7:  "PUBCOMP",
	8:  "SUBSCRIBE",
	9:  "SUBACK",
	10: "UNSUBSCRIBE",
	11: "UNSUBACK",
	12: "PINGREQ",
	13: "PINGRESP",
	14: "DISCONNECT",
	15: "AUTH",
}


var typeDefaultFlags = [AUTH + 1]byte {
	0, // RESERVED
	0, // CONNECT
	0, // CONNACK
	0, // PUBLISH
	0, // PUBACK
	0, // PUBREC
	2, // PUBREL
	0, // PUBCOMP
	2, // SUBSCRIBE
	0, // SUBACK
	2, // UNSUBSCRIBE
	0, // UNSUBACK
	0, // PINGREQ
	0, // PINGRESP
	0, // DISCONNECT
	0, // AUTH
}

func PacketName(t byte) string {
	if t > AUTH {
		return "UNKNOWN"
	}
	return PacketNames[t]
}

// DefaultFlags returns the default flag values for the message type, as defined by the MQTT spec.
func DefaultFlags(t byte) byte {
	if t > AUTH {
		return 0
	}
	return typeDefaultFlags[t]
}
