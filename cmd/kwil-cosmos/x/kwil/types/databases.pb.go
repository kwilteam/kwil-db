// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: kwil/databases.proto

package types

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type Databases struct {
	Index string `protobuf:"bytes,1,opt,name=index,proto3" json:"index,omitempty"`
	Dbid  string `protobuf:"bytes,2,opt,name=dbid,proto3" json:"dbid,omitempty"`
	Owner string `protobuf:"bytes,3,opt,name=owner,proto3" json:"owner,omitempty"`
}

func (m *Databases) Reset()         { *m = Databases{} }
func (m *Databases) String() string { return proto.CompactTextString(m) }
func (*Databases) ProtoMessage()    {}
func (*Databases) Descriptor() ([]byte, []int) {
	return fileDescriptor_d51fd5e8d82b91ea, []int{0}
}
func (m *Databases) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Databases) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Databases.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Databases) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Databases.Merge(m, src)
}
func (m *Databases) XXX_Size() int {
	return m.Size()
}
func (m *Databases) XXX_DiscardUnknown() {
	xxx_messageInfo_Databases.DiscardUnknown(m)
}

var xxx_messageInfo_Databases proto.InternalMessageInfo

func (m *Databases) GetIndex() string {
	if m != nil {
		return m.Index
	}
	return ""
}

func (m *Databases) GetDbid() string {
	if m != nil {
		return m.Dbid
	}
	return ""
}

func (m *Databases) GetOwner() string {
	if m != nil {
		return m.Owner
	}
	return ""
}

func init() {
	proto.RegisterType((*Databases)(nil), "kwil.kwil.Databases")
}

func init() { proto.RegisterFile("kwil/databases.proto", fileDescriptor_d51fd5e8d82b91ea) }

var fileDescriptor_d51fd5e8d82b91ea = []byte{
	// 181 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xc9, 0x2e, 0xcf, 0xcc,
	0xd1, 0x4f, 0x49, 0x2c, 0x49, 0x4c, 0x4a, 0x2c, 0x4e, 0x2d, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9,
	0x17, 0xe2, 0x04, 0x89, 0xea, 0x81, 0x08, 0x25, 0x6f, 0x2e, 0x4e, 0x17, 0x98, 0xac, 0x90, 0x08,
	0x17, 0x6b, 0x66, 0x5e, 0x4a, 0x6a, 0x85, 0x04, 0xa3, 0x02, 0xa3, 0x06, 0x67, 0x10, 0x84, 0x23,
	0x24, 0xc4, 0xc5, 0x92, 0x92, 0x94, 0x99, 0x22, 0xc1, 0x04, 0x16, 0x04, 0xb3, 0x41, 0x2a, 0xf3,
	0xcb, 0xf3, 0x52, 0x8b, 0x24, 0x98, 0x21, 0x2a, 0xc1, 0x1c, 0xa7, 0xa0, 0x13, 0x8f, 0xe4, 0x18,
	0x2f, 0x3c, 0x92, 0x63, 0x7c, 0xf0, 0x48, 0x8e, 0x71, 0xc2, 0x63, 0x39, 0x86, 0x0b, 0x8f, 0xe5,
	0x18, 0x6e, 0x3c, 0x96, 0x63, 0x88, 0xb2, 0x48, 0xcf, 0x2c, 0xc9, 0x28, 0x4d, 0xd2, 0x4b, 0xce,
	0xcf, 0xd5, 0x07, 0xd9, 0x5b, 0x92, 0x9a, 0x08, 0x61, 0xe8, 0xa6, 0x24, 0xe9, 0x27, 0xe7, 0xa6,
	0x40, 0xd8, 0xc9, 0xf9, 0xc5, 0xb9, 0xf9, 0xc5, 0xfa, 0x15, 0x60, 0x9e, 0x7e, 0x49, 0x65, 0x41,
	0x6a, 0x71, 0x12, 0x1b, 0xd8, 0xc9, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0xfa, 0xf5, 0xc9,
	0xaa, 0xca, 0x00, 0x00, 0x00,
}

func (m *Databases) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Databases) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Databases) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Owner) > 0 {
		i -= len(m.Owner)
		copy(dAtA[i:], m.Owner)
		i = encodeVarintDatabases(dAtA, i, uint64(len(m.Owner)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Dbid) > 0 {
		i -= len(m.Dbid)
		copy(dAtA[i:], m.Dbid)
		i = encodeVarintDatabases(dAtA, i, uint64(len(m.Dbid)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Index) > 0 {
		i -= len(m.Index)
		copy(dAtA[i:], m.Index)
		i = encodeVarintDatabases(dAtA, i, uint64(len(m.Index)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintDatabases(dAtA []byte, offset int, v uint64) int {
	offset -= sovDatabases(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Databases) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Index)
	if l > 0 {
		n += 1 + l + sovDatabases(uint64(l))
	}
	l = len(m.Dbid)
	if l > 0 {
		n += 1 + l + sovDatabases(uint64(l))
	}
	l = len(m.Owner)
	if l > 0 {
		n += 1 + l + sovDatabases(uint64(l))
	}
	return n
}

func sovDatabases(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozDatabases(x uint64) (n int) {
	return sovDatabases(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Databases) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDatabases
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Databases: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Databases: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDatabases
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDatabases
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthDatabases
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Index = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Dbid", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDatabases
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDatabases
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthDatabases
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Dbid = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Owner", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDatabases
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthDatabases
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthDatabases
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Owner = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDatabases(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthDatabases
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipDatabases(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowDatabases
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowDatabases
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowDatabases
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthDatabases
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupDatabases
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthDatabases
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthDatabases        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowDatabases          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupDatabases = fmt.Errorf("proto: unexpected end of group")
)
