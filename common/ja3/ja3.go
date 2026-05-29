// Copyright (c) 2018, Open Systems AG. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

package ja3

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
)

type ClientHello struct {
	Version             uint16
	CipherSuites        []uint16
	Extensions          []uint16
	EllipticCurves      []uint16
	EllipticCurvePF     []uint8
	Versions            []uint16
	SignatureAlgorithms []uint16
	ServerName          string
	ja3ByteString       []byte
	ja3Hash             string
}

func (j *ClientHello) Equals(another *ClientHello, ignoreExtensionsSequence bool) bool {
	if j.Version != another.Version {
		return false
	}
	if !equalUint16(j.CipherSuites, another.CipherSuites) {
		return false
	}
	if !ignoreExtensionsSequence && !equalUint16(j.Extensions, another.Extensions) {
		return false
	}
	if ignoreExtensionsSequence && !equalUint16(j.Extensions, another.sortedExtensions()) {
		return false
	}
	if !equalUint16(j.EllipticCurves, another.EllipticCurves) {
		return false
	}
	if !equalUint8(j.EllipticCurvePF, another.EllipticCurvePF) {
		return false
	}
	if !equalUint16(j.SignatureAlgorithms, another.SignatureAlgorithms) {
		return false
	}
	return true
}

func (j *ClientHello) sortedExtensions() []uint16 {
	extensions := make([]uint16, len(j.Extensions))
	copy(extensions, j.Extensions)
	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i] < extensions[j]
	})
	return extensions
}

func equalUint16(left []uint16, right []uint16) bool {
	if len(left) != len(right) {
		return false
	}
	for i, value := range left {
		if value != right[i] {
			return false
		}
	}
	return true
}

func equalUint8(left []uint8, right []uint8) bool {
	if len(left) != len(right) {
		return false
	}
	for i, value := range left {
		if value != right[i] {
			return false
		}
	}
	return true
}

func Compute(payload []byte) (*ClientHello, error) {
	ja3 := ClientHello{}
	err := ja3.parseSegment(payload)
	return &ja3, err
}

func (j *ClientHello) String() string {
	if j.ja3ByteString == nil {
		j.marshalJA3()
	}
	return string(j.ja3ByteString)
}

func (j *ClientHello) Hash() string {
	if j.ja3ByteString == nil {
		j.marshalJA3()
	}
	if j.ja3Hash == "" {
		h := md5.Sum(j.ja3ByteString)
		j.ja3Hash = hex.EncodeToString(h[:])
	}
	return j.ja3Hash
}
