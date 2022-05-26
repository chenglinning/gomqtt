package mqttp

import (
	"github.com/wonderivan/logger"
	"encoding/binary"
	"regexp"
	"unicode/utf8"
	"bytes"
	"io"
	"fmt"
	"time"
)

type PingReq struct {
	header
}

var _ Packet = (*PingReq)(nil)

func NewPingReq() *PingReq {
	p := &PingReq{}
	return p
}

func (this *PingReq) Unpack(rdata []byte) error {
	return nil
}

func (this *PingReq) Pack() ([]byte, error) {
	return []byte{}, nil
}
