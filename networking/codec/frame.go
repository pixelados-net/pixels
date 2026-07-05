package codec

import "encoding/binary"

// AppendFrame appends an encoded packet frame to dst.
func AppendFrame(dst []byte, packet Packet) ([]byte, error) {
	size := packet.Size()
	if uint64(size) > uint64(^uint32(0)) {
		return dst, ErrPayloadTooLarge
	}

	offset := len(dst)
	dst = append(dst, 0, 0, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(dst[offset:offset+LengthSize], uint32(size))
	binary.BigEndian.PutUint16(dst[offset+LengthSize:offset+FrameOverhead], packet.Header)
	dst = append(dst, packet.Payload...)

	return dst, nil
}

// DecodeFrames decodes all complete frames from src into dst.
func DecodeFrames(dst []Packet, src []byte) ([]Packet, []byte, error) {
	for len(src) >= LengthSize {
		size := int(binary.BigEndian.Uint32(src[:LengthSize]))
		if size < HeaderSize {
			return dst, src, ErrFrameTooSmall
		}

		end := LengthSize + size
		if len(src) < end {
			return dst, src, nil
		}

		frame := src[LengthSize:end]
		dst = append(dst, Packet{
			Header:  binary.BigEndian.Uint16(frame[:HeaderSize]),
			Payload: frame[HeaderSize:],
		})
		src = src[end:]
	}

	return dst, src, nil
}
