/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: k8s.io/kubernetes/vendor/k8s.io/api/events/v1/generated.proto

package v1

import (
	fmt "fmt"

	io "io"

	proto "github.com/gogo/protobuf/proto"
	v11 "k8s.io/api/core/v1"

	math "math"
	math_bits "math/bits"
	reflect "reflect"
	strings "strings"
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

func (m *Event) Reset()      { *m = Event{} }
func (*Event) ProtoMessage() {}
func (*Event) Descriptor() ([]byte, []int) {
	return fileDescriptor_ee2600587b650fac, []int{0}
}
func (m *Event) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Event) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *Event) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Event.Merge(m, src)
}
func (m *Event) XXX_Size() int {
	return m.Size()
}
func (m *Event) XXX_DiscardUnknown() {
	xxx_messageInfo_Event.DiscardUnknown(m)
}

var xxx_messageInfo_Event proto.InternalMessageInfo

func (m *EventList) Reset()      { *m = EventList{} }
func (*EventList) ProtoMessage() {}
func (*EventList) Descriptor() ([]byte, []int) {
	return fileDescriptor_ee2600587b650fac, []int{1}
}
func (m *EventList) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *EventList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventList.Merge(m, src)
}
func (m *EventList) XXX_Size() int {
	return m.Size()
}
func (m *EventList) XXX_DiscardUnknown() {
	xxx_messageInfo_EventList.DiscardUnknown(m)
}

var xxx_messageInfo_EventList proto.InternalMessageInfo

func (m *EventSeries) Reset()      { *m = EventSeries{} }
func (*EventSeries) ProtoMessage() {}
func (*EventSeries) Descriptor() ([]byte, []int) {
	return fileDescriptor_ee2600587b650fac, []int{2}
}
func (m *EventSeries) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventSeries) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *EventSeries) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventSeries.Merge(m, src)
}
func (m *EventSeries) XXX_Size() int {
	return m.Size()
}
func (m *EventSeries) XXX_DiscardUnknown() {
	xxx_messageInfo_EventSeries.DiscardUnknown(m)
}

var xxx_messageInfo_EventSeries proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Event)(nil), "k8s.io.api.events.v1.Event")
	proto.RegisterType((*EventList)(nil), "k8s.io.api.events.v1.EventList")
	proto.RegisterType((*EventSeries)(nil), "k8s.io.api.events.v1.EventSeries")
}

func init() {
	proto.RegisterFile("k8s.io/kubernetes/vendor/k8s.io/api/events/v1/generated.proto", fileDescriptor_ee2600587b650fac)
}

var fileDescriptor_ee2600587b650fac = []byte{
	// 772 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x54, 0x4f, 0x6f, 0xe3, 0x44,
	0x14, 0x8f, 0xbb, 0x4d, 0xda, 0x4c, 0x76, 0xb7, 0xe9, 0x2c, 0x52, 0x87, 0xae, 0xe4, 0x84, 0xac,
	0x84, 0x22, 0x24, 0x6c, 0xb2, 0x42, 0x88, 0x0b, 0x12, 0xeb, 0xa6, 0xa0, 0xa2, 0x96, 0x4a, 0xd3,
	0x9e, 0x10, 0x87, 0x4e, 0x9c, 0x57, 0xd7, 0x24, 0x9e, 0xb1, 0x66, 0x26, 0x91, 0x7a, 0xe3, 0x82,
	0xc4, 0x91, 0x2f, 0xc0, 0x07, 0x40, 0x7c, 0x91, 0x1e, 0x7b, 0xec, 0x29, 0xa2, 0xe6, 0x8b, 0x20,
	0x8f, 0x9d, 0x38, 0xcd, 0x1f, 0x08, 0xda, 0x9b, 0xe7, 0xbd, 0xdf, 0x9f, 0xf7, 0x66, 0x9e, 0x1f,
	0xfa, 0x6a, 0xf0, 0xa5, 0x72, 0x42, 0xe1, 0x0e, 0x46, 0x3d, 0x90, 0x1c, 0x34, 0x28, 0x77, 0x0c,
	0xbc, 0x2f, 0xa4, 0x9b, 0x27, 0x58, 0x1c, 0xba, 0x30, 0x06, 0xae, 0x95, 0x3b, 0xee, 0xb8, 0x01,
	0x70, 0x90, 0x4c, 0x43, 0xdf, 0x89, 0xa5, 0xd0, 0x02, 0x7f, 0x90, 0xa1, 0x1c, 0x16, 0x87, 0x4e,
	0x86, 0x72, 0xc6, 0x9d, 0xc3, 0x4f, 0x83, 0x50, 0xdf, 0x8c, 0x7a, 0x8e, 0x2f, 0x22, 0x37, 0x10,
	0x81, 0x70, 0x0d, 0xb8, 0x37, 0xba, 0x36, 0x27, 0x73, 0x30, 0x5f, 0x99, 0xc8, 0x61, 0x6b, 0xce,
	0xca, 0x17, 0x12, 0x56, 0x18, 0x1d, 0x7e, 0x5e, 0x60, 0x22, 0xe6, 0xdf, 0x84, 0x1c, 0xe4, 0xad,
	0x1b, 0x0f, 0x82, 0x34, 0xa0, 0xdc, 0x08, 0x34, 0x5b, 0xc5, 0x72, 0xd7, 0xb1, 0xe4, 0x88, 0xeb,
	0x30, 0x82, 0x25, 0xc2, 0x17, 0xff, 0x45, 0x50, 0xfe, 0x0d, 0x44, 0x6c, 0x91, 0xd7, 0xfa, 0xbd,
	0x8a, 0xca, 0xc7, 0x69, 0xff, 0xf8, 0x0a, 0xed, 0xa6, 0xd5, 0xf4, 0x99, 0x66, 0xc4, 0x6a, 0x5a,
	0xed, 0xda, 0xdb, 0xcf, 0x9c, 0xe2, 0x92, 0x66, 0xa2, 0x4e, 0x3c, 0x08, 0xd2, 0x80, 0x72, 0x52,
	0xb4, 0x33, 0xee, 0x38, 0xe7, 0xbd, 0x9f, 0xc0, 0xd7, 0x67, 0xa0, 0x99, 0x87, 0xef, 0x26, 0x8d,
	0x52, 0x32, 0x69, 0xa0, 0x22, 0x46, 0x67, 0xaa, 0xf8, 0x0a, 0x55, 0xcd, 0x55, 0x5f, 0x86, 0x11,
	0x90, 0x2d, 0x63, 0xe1, 0x6e, 0x66, 0x71, 0x16, 0xfa, 0x52, 0xa4, 0x34, 0x6f, 0x3f, 0x77, 0xa8,
	0x1e, 0x4f, 0x95, 0x68, 0x21, 0x8a, 0x8f, 0x51, 0x45, 0x81, 0x0c, 0x41, 0x91, 0x67, 0x46, 0xfe,
	0x23, 0x67, 0xd5, 0x33, 0x3b, 0x86, 0x7b, 0x61, 0x80, 0x1e, 0x4a, 0x26, 0x8d, 0x4a, 0xf6, 0x4d,
	0x73, 0x32, 0x3e, 0x43, 0xaf, 0x24, 0xc4, 0x42, 0xea, 0x90, 0x07, 0x47, 0x82, 0x6b, 0x29, 0x86,
	0x43, 0x90, 0x64, 0xbb, 0x69, 0xb5, 0xab, 0xde, 0xeb, 0xbc, 0x82, 0x57, 0x74, 0x19, 0x42, 0x57,
	0xf1, 0xf0, 0xb7, 0x68, 0x7f, 0x16, 0x3e, 0xe1, 0x4a, 0x33, 0xee, 0x03, 0x29, 0x1b, 0xb1, 0x0f,
	0x73, 0xb1, 0x7d, 0xba, 0x08, 0xa0, 0xcb, 0x1c, 0xfc, 0x31, 0xaa, 0x30, 0x5f, 0x87, 0x82, 0x93,
	0x8a, 0x61, 0xbf, 0xcc, 0xd9, 0x95, 0x77, 0x26, 0x4a, 0xf3, 0x6c, 0x8a, 0x93, 0xc0, 0x94, 0xe0,
	0x64, 0xe7, 0x29, 0x8e, 0x9a, 0x28, 0xcd, 0xb3, 0xf8, 0x12, 0x55, 0x25, 0x04, 0x4c, 0xf6, 0x43,
	0x1e, 0x90, 0x5d, 0x73, 0x63, 0x6f, 0xe6, 0x6f, 0x2c, 0x9d, 0xe9, 0xe2, 0x85, 0x29, 0x5c, 0x83,
	0x04, 0xee, 0xcf, 0x3d, 0x02, 0x9d, 0xb2, 0x69, 0x21, 0x84, 0xbf, 0x43, 0x3b, 0x12, 0x86, 0xe9,
	0x8c, 0x91, 0xea, 0xe6, 0x9a, 0xb5, 0x64, 0xd2, 0xd8, 0xa1, 0x19, 0x8f, 0x4e, 0x05, 0x70, 0x13,
	0x6d, 0x73, 0xa1, 0x81, 0x20, 0xd3, 0xc7, 0xf3, 0xdc, 0x77, 0xfb, 0x7b, 0xa1, 0x81, 0x9a, 0x4c,
	0x8a, 0xd0, 0xb7, 0x31, 0x90, 0xda, 0x53, 0xc4, 0xe5, 0x6d, 0x0c, 0xd4, 0x64, 0x30, 0xa0, 0x7a,
	0x1f, 0x62, 0x09, 0x7e, 0xaa, 0x78, 0x21, 0x46, 0xd2, 0x07, 0xf2, 0xdc, 0x14, 0xd6, 0x58, 0x55,
	0x58, 0x36, 0x1c, 0x06, 0xe6, 0x91, 0x5c, 0xae, 0xde, 0x5d, 0x10, 0xa0, 0x4b, 0x92, 0xf8, 0x57,
	0x0b, 0x91, 0x22, 0xf8, 0x4d, 0x28, 0x95, 0x99, 0x49, 0xa5, 0x59, 0x14, 0x93, 0x17, 0xc6, 0xef,
	0x93, 0xcd, 0xa6, 0xdd, 0x0c, 0x7a, 0x33, 0xb7, 0x26, 0xdd, 0x35, 0x9a, 0x74, 0xad, 0x1b, 0xfe,
	0xc5, 0x42, 0x07, 0x45, 0xf2, 0x94, 0xcd, 0x57, 0xf2, 0xf2, 0x7f, 0x57, 0xd2, 0xc8, 0x2b, 0x39,
	0xe8, 0xae, 0x96, 0xa4, 0xeb, 0xbc, 0xf0, 0x3b, 0xb4, 0x57, 0xa4, 0x8e, 0xc4, 0x88, 0x6b, 0xb2,
	0xd7, 0xb4, 0xda, 0x65, 0xef, 0x20, 0x97, 0xdc, 0xeb, 0x3e, 0x4d, 0xd3, 0x45, 0x7c, 0xeb, 0x4f,
	0x0b, 0x65, 0xbf, 0xfa, 0x69, 0xa8, 0x34, 0xfe, 0x71, 0x69, 0x47, 0x39, 0x9b, 0x35, 0x92, 0xb2,
	0xcd, 0x86, 0xaa, 0xe7, 0xce, 0xbb, 0xd3, 0xc8, 0xdc, 0x7e, 0xfa, 0x1a, 0x95, 0x43, 0x0d, 0x91,
	0x22, 0x5b, 0xcd, 0x67, 0xed, 0xda, 0xdb, 0xd7, 0xff, 0xb2, 0x3c, 0xbc, 0x17, 0xb9, 0x4e, 0xf9,
	0x24, 0x65, 0xd0, 0x8c, 0xd8, 0xfa, 0xc3, 0x42, 0xb5, 0xb9, 0xe5, 0x82, 0xdf, 0xa0, 0xb2, 0x6f,
	0xda, 0xb6, 0x4c, 0xdb, 0x33, 0x52, 0xd6, 0x6c, 0x96, 0xc3, 0x23, 0x54, 0x1f, 0x32, 0xa5, 0xcf,
	0x7b, 0x0a, 0xe4, 0x18, 0xfa, 0xef, 0xb3, 0x1d, 0x67, 0xf3, 0x7a, 0xba, 0x20, 0x48, 0x97, 0x2c,
	0xbc, 0xf6, 0xdd, 0xa3, 0x5d, 0xba, 0x7f, 0xb4, 0x4b, 0x0f, 0x8f, 0x76, 0xe9, 0xe7, 0xc4, 0xb6,
	0xee, 0x12, 0xdb, 0xba, 0x4f, 0x6c, 0xeb, 0x21, 0xb1, 0xad, 0xbf, 0x12, 0xdb, 0xfa, 0xed, 0x6f,
	0xbb, 0xf4, 0xc3, 0xd6, 0xb8, 0xf3, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x6b, 0xcb, 0x1e, 0x6e,
	0x6b, 0x07, 0x00, 0x00,
}

func (m *Event) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Event) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Event) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	i = encodeVarintGenerated(dAtA, i, uint64(m.DeprecatedCount))
	i--
	dAtA[i] = 0x78
	{
		size, err := m.DeprecatedLastTimestamp.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenerated(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x72
	{
		size, err := m.DeprecatedFirstTimestamp.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenerated(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x6a
	{
		size, err := m.DeprecatedSource.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenerated(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x62
	i -= len(m.Type)
	copy(dAtA[i:], m.Type)
	i = encodeVarintGenerated(dAtA, i, uint64(len(m.Type)))
	i--
	dAtA[i] = 0x5a
	i -= len(m.Note)
	copy(dAtA[i:], m.Note)
	i = encodeVarintGenerated(dAtA, i, uint64(len(m.Note)))
	i--
	dAtA[i] = 0x52
	if m.Related != nil {
		{
			size, err := m.Related.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintGenerated(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x4a
	}
	{
		size, err := m.Regarding.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenerated(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x42
	i -= len(m.Reason)
	copy(dAtA[i:], m.Reason)
	i = encodeVarintGenerated(dAtA, i, uint64(len(m.Reason)))
	i--
	dAtA[i] = 0x3a
	i -= len(m.Action)
	copy(dAtA[i:], m.Action)
	i = encodeVarintGenerated(dAtA, i, uint64(len(m.Action)))
	i--
	dAtA[i] = 0x32
	i -= len(m.ReportingInstance)
	copy(dAtA[i:], m.ReportingInstance)
	i = encodeVarintGenerated(dAtA, i, uint64(len(m.ReportingInstance)))
	i--
	dAtA[i] = 0x2a
	i -= len(m.ReportingController)
	copy(dAtA[i:], m.ReportingController)
	i = encodeVarintGenerated(dAtA, i, uint64(len(m.ReportingController)))
	i--
	dAtA[i] = 0x22
	if m.Series != nil {
		{
			size, err := m.Series.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintGenerated(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	{
		size, err := m.EventTime.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenerated(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	{
		size, err := m.ObjectMeta.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenerated(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *EventList) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventList) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventList) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Items) > 0 {
		for iNdEx := len(m.Items) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Items[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenerated(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	{
		size, err := m.ListMeta.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenerated(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *EventSeries) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventSeries) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventSeries) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.LastObservedTime.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenerated(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	i = encodeVarintGenerated(dAtA, i, uint64(m.Count))
	i--
	dAtA[i] = 0x8
	return len(dAtA) - i, nil
}

func encodeVarintGenerated(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenerated(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Event) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.ObjectMeta.Size()
	n += 1 + l + sovGenerated(uint64(l))
	l = m.EventTime.Size()
	n += 1 + l + sovGenerated(uint64(l))
	if m.Series != nil {
		l = m.Series.Size()
		n += 1 + l + sovGenerated(uint64(l))
	}
	l = len(m.ReportingController)
	n += 1 + l + sovGenerated(uint64(l))
	l = len(m.ReportingInstance)
	n += 1 + l + sovGenerated(uint64(l))
	l = len(m.Action)
	n += 1 + l + sovGenerated(uint64(l))
	l = len(m.Reason)
	n += 1 + l + sovGenerated(uint64(l))
	l = m.Regarding.Size()
	n += 1 + l + sovGenerated(uint64(l))
	if m.Related != nil {
		l = m.Related.Size()
		n += 1 + l + sovGenerated(uint64(l))
	}
	l = len(m.Note)
	n += 1 + l + sovGenerated(uint64(l))
	l = len(m.Type)
	n += 1 + l + sovGenerated(uint64(l))
	l = m.DeprecatedSource.Size()
	n += 1 + l + sovGenerated(uint64(l))
	l = m.DeprecatedFirstTimestamp.Size()
	n += 1 + l + sovGenerated(uint64(l))
	l = m.DeprecatedLastTimestamp.Size()
	n += 1 + l + sovGenerated(uint64(l))
	n += 1 + sovGenerated(uint64(m.DeprecatedCount))
	return n
}

func (m *EventList) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.ListMeta.Size()
	n += 1 + l + sovGenerated(uint64(l))
	if len(m.Items) > 0 {
		for _, e := range m.Items {
			l = e.Size()
			n += 1 + l + sovGenerated(uint64(l))
		}
	}
	return n
}

func (m *EventSeries) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 1 + sovGenerated(uint64(m.Count))
	l = m.LastObservedTime.Size()
	n += 1 + l + sovGenerated(uint64(l))
	return n
}

func sovGenerated(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenerated(x uint64) (n int) {
	return sovGenerated(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *Event) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Event{`,
		`ObjectMeta:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.ObjectMeta), "ObjectMeta", "v1.ObjectMeta", 1), `&`, ``, 1) + `,`,
		`EventTime:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.EventTime), "MicroTime", "v1.MicroTime", 1), `&`, ``, 1) + `,`,
		`Series:` + strings.Replace(this.Series.String(), "EventSeries", "EventSeries", 1) + `,`,
		`ReportingController:` + fmt.Sprintf("%v", this.ReportingController) + `,`,
		`ReportingInstance:` + fmt.Sprintf("%v", this.ReportingInstance) + `,`,
		`Action:` + fmt.Sprintf("%v", this.Action) + `,`,
		`Reason:` + fmt.Sprintf("%v", this.Reason) + `,`,
		`Regarding:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.Regarding), "ObjectReference", "v11.ObjectReference", 1), `&`, ``, 1) + `,`,
		`Related:` + strings.Replace(fmt.Sprintf("%v", this.Related), "ObjectReference", "v11.ObjectReference", 1) + `,`,
		`Note:` + fmt.Sprintf("%v", this.Note) + `,`,
		`Type:` + fmt.Sprintf("%v", this.Type) + `,`,
		`DeprecatedSource:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.DeprecatedSource), "EventSource", "v11.EventSource", 1), `&`, ``, 1) + `,`,
		`DeprecatedFirstTimestamp:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.DeprecatedFirstTimestamp), "Time", "v1.Time", 1), `&`, ``, 1) + `,`,
		`DeprecatedLastTimestamp:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.DeprecatedLastTimestamp), "Time", "v1.Time", 1), `&`, ``, 1) + `,`,
		`DeprecatedCount:` + fmt.Sprintf("%v", this.DeprecatedCount) + `,`,
		`}`,
	}, "")
	return s
}
func (this *EventList) String() string {
	if this == nil {
		return "nil"
	}
	repeatedStringForItems := "[]Event{"
	for _, f := range this.Items {
		repeatedStringForItems += strings.Replace(strings.Replace(f.String(), "Event", "Event", 1), `&`, ``, 1) + ","
	}
	repeatedStringForItems += "}"
	s := strings.Join([]string{`&EventList{`,
		`ListMeta:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.ListMeta), "ListMeta", "v1.ListMeta", 1), `&`, ``, 1) + `,`,
		`Items:` + repeatedStringForItems + `,`,
		`}`,
	}, "")
	return s
}
func (this *EventSeries) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&EventSeries{`,
		`Count:` + fmt.Sprintf("%v", this.Count) + `,`,
		`LastObservedTime:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.LastObservedTime), "MicroTime", "v1.MicroTime", 1), `&`, ``, 1) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringGenerated(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *Event) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenerated
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
			return fmt.Errorf("proto: Event: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Event: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ObjectMeta", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ObjectMeta.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EventTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.EventTime.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Series", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Series == nil {
				m.Series = &EventSeries{}
			}
			if err := m.Series.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReportingController", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
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
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ReportingController = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReportingInstance", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
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
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ReportingInstance = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Action", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
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
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Action = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Reason", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
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
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Reason = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Regarding", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Regarding.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Related", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Related == nil {
				m.Related = &v11.ObjectReference{}
			}
			if err := m.Related.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Note", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
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
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Note = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 11:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
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
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Type = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 12:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DeprecatedSource", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.DeprecatedSource.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 13:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DeprecatedFirstTimestamp", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.DeprecatedFirstTimestamp.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 14:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DeprecatedLastTimestamp", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.DeprecatedLastTimestamp.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 15:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field DeprecatedCount", wireType)
			}
			m.DeprecatedCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.DeprecatedCount |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipGenerated(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenerated
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
func (m *EventList) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenerated
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
			return fmt.Errorf("proto: EventList: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventList: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ListMeta", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ListMeta.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Items", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Items = append(m.Items, Event{})
			if err := m.Items[len(m.Items)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenerated(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenerated
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
func (m *EventSeries) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenerated
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
			return fmt.Errorf("proto: EventSeries: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventSeries: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Count", wireType)
			}
			m.Count = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Count |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastObservedTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.LastObservedTime.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenerated(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenerated
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
func skipGenerated(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenerated
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
					return 0, ErrIntOverflowGenerated
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
					return 0, ErrIntOverflowGenerated
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
				return 0, ErrInvalidLengthGenerated
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenerated
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenerated
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenerated        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenerated          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenerated = fmt.Errorf("proto: unexpected end of group")
)
