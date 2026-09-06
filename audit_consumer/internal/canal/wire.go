package canal

import (
	"errors"
	"fmt"
)

var (
	errTruncated   = errors.New("protobuf 字节流截断")
	errBadVarint   = errors.New("protobuf varint 编码非法")
	errWireType    = errors.New("不支持的 protobuf wire type")
	errCompression = errors.New("不支持的报文压缩方式")
)

// pbReader 仅处理 protobuf wire 格式；它不包含任何 Canal 业务语义。
type pbReader struct {
	b   []byte
	off int
}

func (r *pbReader) done() bool { return r.off >= len(r.b) }

func (r *pbReader) varint() (uint64, error) {
	var value uint64
	var shift uint
	for index := 0; index < 10; index++ {
		if r.off >= len(r.b) {
			return 0, errTruncated
		}
		current := r.b[r.off]
		r.off++
		if current < 0x80 {
			if index == 9 && current > 1 {
				return 0, errBadVarint
			}
			return value | uint64(current)<<shift, nil
		}
		value |= uint64(current&0x7f) << shift
		shift += 7
	}
	return 0, errBadVarint
}

func (r *pbReader) bytes() ([]byte, error) {
	length, err := r.varint()
	if err != nil {
		return nil, err
	}
	if length > uint64(len(r.b)-r.off) {
		return nil, errTruncated
	}
	value := r.b[r.off : r.off+int(length)]
	r.off += int(length)
	return value, nil
}

func (r *pbReader) skip(wire int) error {
	switch wire {
	case 0:
		_, err := r.varint()
		return err
	case 1:
		if len(r.b)-r.off < 8 {
			return errTruncated
		}
		r.off += 8
		return nil
	case 2:
		_, err := r.bytes()
		return err
	case 5:
		if len(r.b)-r.off < 4 {
			return errTruncated
		}
		r.off += 4
		return nil
	default:
		return fmt.Errorf("%w: %d", errWireType, wire)
	}
}

func (r *pbReader) nextField() (int, int, error) {
	tag, err := r.varint()
	if err != nil {
		return 0, 0, err
	}
	return int(tag >> 3), int(tag & 7), nil
}
