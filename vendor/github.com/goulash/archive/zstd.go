package archive

import (
	"io"

	"github.com/klauspost/compress/zstd"
)

// zstDecompressor wraps the zstd.Decoder type to implement io.Closer,
// which it unfortunately doesn't quite implement.
type zstDecompressor struct {
	decoder *zstd.Decoder
}

func newZstDecompressor(r io.Reader) (*zstDecompressor, error) {
	var d zstDecompressor
	z, err := zstd.NewReader(r)
	if err != nil {
		return nil, err
	}
	d.decoder = z
	return &d, nil
}

func (d *zstDecompressor) Read(p []byte) (int, error) {
	return d.decoder.Read(p)
}

func (d *zstDecompressor) Close() error {
	d.decoder.Close()
	return nil
}
