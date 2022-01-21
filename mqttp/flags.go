package mqttp

const (
	maskMessageFlags               byte = 0x0F
	maskConnFlagUsername           byte = 0x80
	maskConnFlagPassword           byte = 0x40
	maskConnFlagWillRetain         byte = 0x20
	maskConnFlagWillQos            byte = 0x18
	maskConnFlagWill               byte = 0x04
	maskConnFlagClean              byte = 0x02
	maskConnFlagReserved           byte = 0x01
	maskPublishFlagRetain          byte = 0x01
	maskPublishFlagQoS             byte = 0x06
	maskPublishFlagDup             byte = 0x08
	maskSubscriptionQoS            byte = 0x03
	maskSubscriptionNL             byte = 0x04
	maskSubscriptionRAP            byte = 0x08
	maskSubscriptionRetainHandling byte = 0x30
	maskSubscriptionReservedV3     byte = 0xFC
	maskSubscriptionReservedV5     byte = 0xC0
)

const (
	maskType   byte = 0xF0
	maskFlags  byte = 0x0F
	maskQoS    byte = 0x06
	maskDup    byte = 0x08
	maskRetain byte = 0x01
)
