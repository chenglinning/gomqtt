package mqttp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)
const (
	MAX_UINT uint64 = 268435455
)
/*
func ReadUvarint(r io.Reader) (uint32, error) {
	var b byte
	var x uint32
	var shift uint32

	buff := make([]byte, 1)

	for i:=0; i<4; i++ {
		_, err := io.ReadFull(r, buff)
		if err != nil {
			return 0, err
		}
		b = buff[0]
		if b < 0x80 {
			return x | uint32(b) << shift, nil
		}
		x |= uint32(b & 0x7f) << shift
		shift += 7
	}
	
	return 0, errors.New("uvarint32 overflow")
}
*/

func ReadByte(r io.Reader) (byte, error) {
	buff := make([]byte, 1)
	_, err := io.ReadFull(r, buff)
	if err != nil {
		return 0, err
	}
	return buff[0], nil
 }

 func WriteByte(w io.Writer, v byte) error {
	buff := []byte{v}
	_, err := w.Write(buff)
	return err
}

 func ReadUint16(r io.Reader) (uint16, error) {
	buff := make([]byte, 2)
	_, err := io.ReadFull(r, buff)
	if err != nil {
		return 0, err
	}
	v := binary.BigEndian.Uint16(buff)
	return v, nil
 }

 func WriteUint16(w io.Writer, v uint16) error {
	buff := make([]byte, 2)
	binary.BigEndian.PutUint16(buff, v)
	_, err := w.Write(buff)
	return err
}
 
 func ReadUint32(r io.Reader) (uint32, error) {
	buff := make([]byte, 4)
	_, err := io.ReadFull(r, buff)
	if err != nil {
		return 0, err
	}
	v := binary.BigEndian.Uint32(buff)
	return v, nil
 }

 func WriteUint32(w io.Writer, v uint32) error {
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, v)
	_, err := w.Write(buff)
	return err
}

 func ReadSting(r io.Reader) (string, error) {
	n, err := ReadUint16(r)
	if err != nil {
		return nil, err
	}
	buff := make([]byte, n)
	_, err := io.ReadFull(r, buff)
	if err != nil {
		return nil, err
	}
	return string(v), nil
 }

 func WriteString(w io.Writer, v string) error {
	buff := []byte(v)
	n := (len(buff))
	err := WriteUint16(w, uint16(n))
	if err != nil {
		return err
	}
	_, err := w.Write(buff)
	return err
}
 
 func ReadStingPair(r io.Reader) (StringPair, error) {
	key, err := ReadSting(r)
	if err != nil {
		return nil, err
	}
	val, err := ReadSting(r)
	if err != nil {
		return nil, err
	}
	return StringPair{k:key, v:val}, nil
 }

 func WriteStringPair(w io.Writer, sp StringPair) error {
	err := WriteString(w, sp.k)
	if err != nil {
		return err
	}
	err = WriteString(w, sp.v)
	if err != nil {
		return err
	}
	return nil
}
  
func ReadUvarint(r io.Reader) (uint32, error) {
   u64, err := binary.ReadUvarint(r)
   if err != nil {
		return 0, err
   }
   if u64 > MAX_UINT {
    	return 0, errors.New("uvarint32 overflow")
   }
   return uint32(u64), nil
}

func WriteUvarint(w io.Writer, n uint32) error {
	buff := make([]byte, 8)
	if n > MAX_UINT {
    	return errors.New("uvarint32 overflow > 268435455")
	}
	m := binary.PutUvarint(buff, uint64(n))
	m, err := w.Write(buff[:m])
	return err
}

func ReadBinaryData(r io.Reader) ([]byte, error) {
	n, err := ReadUint16(r)
	if err != nil {
		return nil, err
	}
	buff := make([]byte, n)
	_, err := io.ReadFull(r, buff)
	if err != nil {
		return nil, err
	}
	return buff, nil
 }

 func WriteBinaryData(w io.Writer, buff []byte) error {
	n := (len(buff))
	err := WriteUint16(w, uint16(n))
	if err != nil {
		return err
	}
	_, err := w.Write(buff)
	return err
}

func boolToByte(b bool) byte {
	switch b {
	case true:
		return 0x01
	default:
		return 0x00
	}
}

func vlen(u uint32) int {
	if u < 128 {
		return 1
	} 
	if u < 16384 {
		return 2
	} 
	if u < 2097152 {
		return 3
	} 
	if u < 268435456 {
		return 4
	}
	return 0
}
