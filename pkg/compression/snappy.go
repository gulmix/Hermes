package compression

import (
	"io"
	"sync"

	"github.com/golang/snappy"
	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCompressor(&snappyCompressor{})
}

type snappyCompressor struct {
	pool sync.Pool
}

func (s *snappyCompressor) Name() string { return Snappy }

func (s *snappyCompressor) Compress(w io.Writer) (io.WriteCloser, error) {
	sw, ok := s.pool.Get().(*snappyWriter)
	if !ok {
		return &snappyWriter{Writer: snappy.NewBufferedWriter(w), pool: &s.pool}, nil
	}
	sw.Reset(w)
	return sw, nil
}

func (s *snappyCompressor) Decompress(r io.Reader) (io.Reader, error) {
	return snappy.NewReader(r), nil
}

type snappyWriter struct {
	*snappy.Writer
	pool *sync.Pool
}

func (sw *snappyWriter) Close() error {
	err := sw.Writer.Close()
	sw.pool.Put(sw)
	return err
}
