package codec

import "time"

type Type string

const (
	GobCodecType  Type = "application/gob"
	JsonCodecType Type = "application/json"
)
const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber    int
	CodecType      Type
	ConnectTimeout time.Duration
}

func (o *Option) IsGeeRpc() bool {
	return o.MagicNumber == MagicNumber
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      GobCodecType,
	ConnectTimeout: time.Second * 10,
}
