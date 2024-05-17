package codec

import (
	"bufio"
	"encoding/gob"
	"io"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

func (g *GobCodec) ReadHeader(h *Header) error {
	return g.dec.Decode(h)
}
func (g *GobCodec) ReadBody(body interface{}) error {
	return g.dec.Decode(body)
}
func (g *GobCodec) Write(header *Header, body interface{}) (err error) {
	defer func() {
		_ = g.buf.Flush()
		if err != nil {
			_ = g.Close()
		}
	}()
	if err = g.enc.Encode(header); err != nil {
		return
	}
	if err = g.enc.Encode(body); err != nil {
		return
	}
	return
}
func (g *GobCodec) Close() error {
	return g.conn.Close()
}
func newGobCodec(coon io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(coon)
	return &GobCodec{
		conn: coon,
		buf:  buf,
		dec:  gob.NewDecoder(coon),
		enc:  gob.NewEncoder(buf),
	}
}
