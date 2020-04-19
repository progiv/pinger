package internal

import (
	"bytes"
	"encoding/binary"
	"time"
)

type pingMessageBody struct {
	time time.Time
}

func (p pingMessageBody) BinaryMarshall() ([]byte, error) {
	var buf bytes.Buffer
	timeBinary, _ := p.time.MarshalBinary()
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(timeBinary)))
	err := binary.Write(&buf, binary.BigEndian, timeBinary)
	return buf.Bytes(), err
}

func (p *pingMessageBody) UnmarshalBinary(data []byte) (err error) {
	buf := bytes.NewReader(data)
	var size uint32
	if err = binary.Read(buf, binary.BigEndian, &size); err != nil {
		return
	}
	timeData := make([]byte, size)
	if err = binary.Read(buf, binary.BigEndian, timeData); err != nil {
		return
	}
	err = p.time.UnmarshalBinary(timeData)
	return
}
