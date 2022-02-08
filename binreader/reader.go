package binreader

import (
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"math"

	"github.com/google/uuid"
)

var (
	ErrNegativeAmount = errors.New("negetive bytes count")
)

func NewReader(r io.ReadSeeker, order binary.ByteOrder) *reader {
	return &reader{
		r:     r,
		order: order,
	}
}

type reader struct {
	r     io.ReadSeeker
	order binary.ByteOrder
}

func (r *reader) ReadAll() ([]byte, error) {
	b, err := ioutil.ReadAll(r.r)
	return b, err
}

func (r *reader) ReadBytes(count int) (n int, out []byte, err error) {
	if count < 0 {
		return 0, nil, ErrNegativeAmount
	}

	if count == 0 {
		return 0, []byte{}, nil
	}

	out = make([]byte, count)
	n, err = io.ReadFull(r.r, out)

	return
}

func (r *reader) ReadUint8() (uint8, error) {
	_, b, err := r.ReadBytes(1)
	if err != nil {
		return 0, err
	}

	return b[0], nil
}

func (r *reader) ReadBool() (bool, error) {
	b, err := r.ReadByte()
	return b != 0, err
}

func (r *reader) ReadByte() (byte, error) {
	return r.ReadUint8()
}

func (r *reader) ReadUint16() (uint16, error) {
	_, b, err := r.ReadBytes(2)
	if err != nil {
		return 0, err
	}

	return r.order.Uint16(b), nil
}

func (r *reader) ReadUint32() (uint32, error) {
	_, b, err := r.ReadBytes(4)
	if err != nil {
		return 0, err
	}

	return r.order.Uint32(b), nil
}

func (r *reader) ReadUint64() (uint64, error) {
	_, b, err := r.ReadBytes(8)
	if err != nil {
		return 0, err
	}

	return r.order.Uint64(b), nil
}

func (r *reader) ReadInt8() (int8, error) {
	i, err := r.ReadUint8()
	return int8(i), err
}

func (r *reader) ReadInt16() (int16, error) {
	i, err := r.ReadUint16()
	return int16(i), err
}

func (r *reader) ReadInt32() (int32, error) {
	i, err := r.ReadUint32()
	return int32(i), err
}

func (r *reader) ReadInt64() (int64, error) {
	i, err := r.ReadUint64()
	return int64(i), err
}

func (r *reader) ReadFloat32() (float32, error) {
	b, err := r.ReadUint32()
	if err != nil {
		return 0, err
	}

	return math.Float32frombits(b), nil
}

func (r *reader) ReadFloat64() (float64, error) {
	b, err := r.ReadUint64()
	if err != nil {
		return 0, err
	}

	return math.Float64frombits(b), nil
}

func (r *reader) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

func (r *reader) Seek(offset int64, whence int) (int64, error) {
	i, err := r.r.Seek(offset, whence)
	return i, err
}

func (r *reader) Peek(n int) ([]byte, error) {
	bytesRead, b, err := r.ReadBytes(n)
	if err != nil {
		return nil, err
	}

	_, err = r.Seek(int64(-bytesRead), io.SeekCurrent) // go back
	return b, err
}

// reads a FString (null-terminated string starting with the length) from r
func (r *reader) ReadFString() (string, error) {
	size, err := r.ReadUint32()
	if err != nil || size == 0 {
		return "", err
	}

	_, buf, err := r.ReadBytes(int(size))
	if err != nil {
		return "", err
	}
	if buf[len(buf)-1] != 0x0 { // ensure it's null-terminated
		return "", errors.New("string is not null terminated")
	}
	return string(buf[:len(buf)-1]), nil // avoid the null charecter while returning
}

// read an array of FStrings. they start wtih the length then the data
func (r *reader) ReadFStringArray() (out []string, err error) {
	size, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}

	for i := uint32(0); i < size; i++ {
		fstr, err := r.ReadFString()
		if err != nil {
			return nil, err
		}
		out = append(out, fstr)
	}

	return
}

// reads a GUID which is stored as 4 uint32 segments written in Big Endian
func (r *reader) ReadGUID() (guid uuid.UUID, err error) {
	data := make([]uint32, 4)
	err = binary.Read(r, binary.BigEndian, &data)
	if err != nil {
		return uuid.Nil, err
	}
	for i, v := range data {
		binary.LittleEndian.PutUint32(guid[i*4:(i+1)*4], v)
	}
	return
}
