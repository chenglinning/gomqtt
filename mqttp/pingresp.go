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

type PingResp struct {
	header
}

var _ Packet = (*PingResp)(nil)

func NewPingReq() *PingResp {
	p := &PingResp{}
	return p
}

func (this *PingReq) Unpack(rdata []byte) error {
	return nil
}

func (this *PingReq) Pack() ([]byte, error) {
	return []byte{}, nil
}
