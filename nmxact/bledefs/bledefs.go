/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package bledefs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const BLE_ATT_ATTR_MAX_LEN = 512

const NmpPlainSvcUuid = "8D53DC1D-1DB7-4CD3-868B-8A527460AA84"
const NmpPlainChrUuid = "DA2E7828-FBCE-4E01-AE9E-261174997C48"
const OmpSvcUuid = 0x9923
const OmpReqChrUuid = "AD7B334F-4637-4B86-90B6-9D787F03D218"
const OmpRspChrUuid = "E9241982-4580-42C4-8831-95048216B256"

type BleAddrType int

const (
	BLE_ADDR_TYPE_PUBLIC  BleAddrType = 0
	BLE_ADDR_TYPE_RANDOM              = 1
	BLE_ADDR_TYPE_RPA_PUB             = 2
	BLE_ADDR_TYPE_RPA_RND             = 3
)

var BleAddrTypeStringMap = map[BleAddrType]string{
	BLE_ADDR_TYPE_PUBLIC:  "public",
	BLE_ADDR_TYPE_RANDOM:  "random",
	BLE_ADDR_TYPE_RPA_PUB: "rpa_pub",
	BLE_ADDR_TYPE_RPA_RND: "rpa_rnd",
}

func BleAddrTypeToString(addrType BleAddrType) string {
	s := BleAddrTypeStringMap[addrType]
	if s == "" {
		return "???"
	}

	return s
}

func BleAddrTypeFromString(s string) (BleAddrType, error) {
	for addrType, name := range BleAddrTypeStringMap {
		if s == name {
			return addrType, nil
		}
	}

	return BleAddrType(0), fmt.Errorf("Invalid BleAddrType string: %s", s)
}

func (a BleAddrType) MarshalJSON() ([]byte, error) {
	return json.Marshal(BleAddrTypeToString(a))
}

func (a *BleAddrType) UnmarshalJSON(data []byte) error {
	var err error

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	*a, err = BleAddrTypeFromString(s)
	return err
}

type BleAddr struct {
	Bytes [6]byte
}

func ParseBleAddr(s string) (BleAddr, error) {
	ba := BleAddr{}

	toks := strings.Split(strings.ToLower(s), ":")
	if len(toks) != 6 {
		return ba, fmt.Errorf("invalid BLE addr string: %s", s)
	}

	for i, t := range toks {
		u64, err := strconv.ParseUint(t, 16, 8)
		if err != nil {
			return ba, err
		}
		ba.Bytes[i] = byte(u64)
	}

	return ba, nil
}

func (ba *BleAddr) String() string {
	var buf bytes.Buffer
	buf.Grow(len(ba.Bytes) * 3)

	for i, b := range ba.Bytes {
		if i != 0 {
			buf.WriteString(":")
		}
		fmt.Fprintf(&buf, "%02x", b)
	}

	return buf.String()
}

func (ba *BleAddr) MarshalJSON() ([]byte, error) {
	return json.Marshal(ba.String())
}

func (ba *BleAddr) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	var err error
	*ba, err = ParseBleAddr(s)
	if err != nil {
		return err
	}

	return nil
}

type BleDev struct {
	AddrType BleAddrType
	Addr     BleAddr
}

func (bd *BleDev) String() string {
	return fmt.Sprintf("%s,%s",
		BleAddrTypeToString(bd.AddrType),
		bd.Addr.String())
}

type BleUuid struct {
	// Set to 0 if the 128-bit UUID should be used.
	Uuid16 uint16

	Uuid128 [16]byte
}

func (bu *BleUuid) String() string {
	if bu.Uuid16 != 0 {
		return fmt.Sprintf("0x%04x", bu.Uuid16)
	}

	var buf bytes.Buffer
	buf.Grow(len(bu.Uuid128)*2 + 3)

	// XXX: For now, only support 128-bit UUIDs.

	for i, b := range bu.Uuid128 {
		switch i {
		case 4, 6, 8, 10:
			buf.WriteString("-")
		}

		fmt.Fprintf(&buf, "%02x", b)
	}

	return buf.String()
}

func ParseUuid(uuidStr string) (BleUuid, error) {
	bu := BleUuid{}

	// First, try to parse as a 16-bit UUID.
	val, err := strconv.ParseUint(uuidStr, 0, 16)
	if err == nil {
		bu.Uuid16 = uint16(val)
		return bu, nil
	}

	// Try to parse as a 128-bit UUID.
	if len(uuidStr) != 36 {
		return bu, fmt.Errorf("Invalid UUID: %s", uuidStr)
	}

	boff := 0
	for i := 0; i < 36; {
		switch i {
		case 8, 13, 18, 23:
			if uuidStr[i] != '-' {
				return bu, fmt.Errorf("Invalid UUID: %s", uuidStr)
			}
			i++

		default:
			u64, err := strconv.ParseUint(uuidStr[i:i+2], 16, 8)
			if err != nil {
				return bu, fmt.Errorf("Invalid UUID: %s", uuidStr)
			}
			bu.Uuid128[boff] = byte(u64)
			i += 2
			boff++
		}
	}

	return bu, nil
}

func (bu *BleUuid) MarshalJSON() ([]byte, error) {
	if bu.Uuid16 != 0 {
		return json.Marshal(bu.Uuid16)
	} else {
		return json.Marshal(bu.String())
	}
}

func (bu *BleUuid) UnmarshalJSON(data []byte) error {
	// First, try a 16-bit UUID.
	if err := json.Unmarshal(data, &bu.Uuid16); err == nil {
		return nil
	}

	// Next, try a 128-bit UUID.
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	var err error
	*bu, err = ParseUuid(s)
	if err != nil {
		return err
	}

	return nil
}

func CompareUuids(a BleUuid, b BleUuid) int {
	if a.Uuid16 != 0 || b.Uuid16 != 0 {
		return int(a.Uuid16) - int(b.Uuid16)
	} else {
		return bytes.Compare(a.Uuid128[:], b.Uuid128[:])
	}
}

type BleScanFilterPolicy int

const (
	BLE_SCAN_FILT_NO_WL        BleScanFilterPolicy = 0
	BLE_SCAN_FILT_USE_WL                           = 1
	BLE_SCAN_FILT_NO_WL_INITA                      = 2
	BLE_SCAN_FILT_USE_WL_INITA                     = 3
)

var BleScanFilterPolicyStringMap = map[BleScanFilterPolicy]string{
	BLE_SCAN_FILT_NO_WL:        "no_wl",
	BLE_SCAN_FILT_USE_WL:       "use_wl",
	BLE_SCAN_FILT_NO_WL_INITA:  "no_wl_inita",
	BLE_SCAN_FILT_USE_WL_INITA: "use_wl_inita",
}

func BleScanFilterPolicyToString(filtPolicy BleScanFilterPolicy) string {
	s := BleScanFilterPolicyStringMap[filtPolicy]
	if s == "" {
		return "???"
	}

	return s
}

func BleScanFilterPolicyFromString(s string) (BleScanFilterPolicy, error) {
	for filtPolicy, name := range BleScanFilterPolicyStringMap {
		if s == name {
			return filtPolicy, nil
		}
	}

	return BleScanFilterPolicy(0),
		fmt.Errorf("Invalid BleScanFilterPolicy string: %s", s)
}

func (a BleScanFilterPolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(BleScanFilterPolicyToString(a))
}

func (a *BleScanFilterPolicy) UnmarshalJSON(data []byte) error {
	var err error

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	*a, err = BleScanFilterPolicyFromString(s)
	return err
}

type BleAdvEventType int

const (
	BLE_ADV_EVENT_IND           BleAdvEventType = 0
	BLE_ADV_EVENT_DIRECT_IND_HD                 = 1
	BLE_ADV_EVENT_SCAN_IND                      = 2
	BLE_ADV_EVENT_NONCONN_IND                   = 3
	BLE_ADV_EVENT_DIRECT_IND_LD                 = 4
)

var BleAdvEventTypeStringMap = map[BleAdvEventType]string{
	BLE_ADV_EVENT_IND:           "ind",
	BLE_ADV_EVENT_DIRECT_IND_HD: "direct_ind_hd",
	BLE_ADV_EVENT_SCAN_IND:      "scan_ind",
	BLE_ADV_EVENT_NONCONN_IND:   "nonconn_ind",
	BLE_ADV_EVENT_DIRECT_IND_LD: "direct_ind_ld",
}

func BleAdvEventTypeToString(advEventType BleAdvEventType) string {
	s := BleAdvEventTypeStringMap[advEventType]
	if s == "" {
		return "???"
	}

	return s
}

func BleAdvEventTypeFromString(s string) (BleAdvEventType, error) {
	for advEventType, name := range BleAdvEventTypeStringMap {
		if s == name {
			return advEventType, nil
		}
	}

	return BleAdvEventType(0),
		fmt.Errorf("Invalid BleAdvEventType string: %s", s)
}

func (a BleAdvEventType) MarshalJSON() ([]byte, error) {
	return json.Marshal(BleAdvEventTypeToString(a))
}

func (a *BleAdvEventType) UnmarshalJSON(data []byte) error {
	var err error

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	*a, err = BleAdvEventTypeFromString(s)
	return err
}

type BleAdvReport struct {
	// These fields are always present.
	EventType BleAdvEventType
	Sender    BleDev
	Rssi      int8
	Data      []byte

	// These fields are only present if the sender included them in its
	// advertisement.
	Flags               uint8     // 0 if not present.
	Uuids16             []uint16  // nil if not present
	Uuids16IsComplete   bool      // false if not present
	Uuids32             []uint32  // false if not present
	Uuids32IsComplete   bool      // false if not present
	Uuids128            []BleUuid // false if not present
	Uuids128IsComplete  bool      // false if not present
	Name                string    // "" if not present.
	NameIsComplete      bool      // false if not present.
	TxPwrLvl            int8      // Check TxPwrLvlIsPresent
	TxPwrLvlIsPresent   bool      // false if not present
	SlaveItvlMin        uint16    // Check SlaveItvlIsPresent
	SlaveItvlMax        uint16    // Check SlaveItvlIsPresent
	SlaveItvlIsPresent  bool      // false if not present
	SvcDataUuid16       []byte    // false if not present
	PublicTgtAddrs      []BleAddr // false if not present
	Appearance          uint16    // Check AppearanceIsPresent
	AppearanceIsPresent bool      // false if not present
	AdvItvl             uint16    // Check AdvItvlIsPresent
	AdvItvlIsPresent    bool      // false if not present
	SvcDataUuid32       []byte    // false if not present
	SvcDataUuid128      []byte    // false if not present
	Uri                 []byte    // false if not present
	MfgData             []byte    // false if not present
}

type BleAdvRptFn func(r BleAdvReport)
type BleAdvPredicate func(adv BleAdvReport) bool

type BleConnDesc struct {
	ConnHandle      uint16
	OwnIdAddrType   BleAddrType
	OwnIdAddr       BleAddr
	OwnOtaAddrType  BleAddrType
	OwnOtaAddr      BleAddr
	PeerIdAddrType  BleAddrType
	PeerIdAddr      BleAddr
	PeerOtaAddrType BleAddrType
	PeerOtaAddr     BleAddr
}

func (d *BleConnDesc) String() string {
	return fmt.Sprintf("conn_handle=%d "+
		"own_id_addr=%s,%s own_ota_addr=%s,%s "+
		"peer_id_addr=%s,%s peer_ota_addr=%s,%s",
		d.ConnHandle,
		BleAddrTypeToString(d.OwnIdAddrType),
		d.OwnIdAddr.String(),
		BleAddrTypeToString(d.OwnOtaAddrType),
		d.OwnOtaAddr.String(),
		BleAddrTypeToString(d.PeerIdAddrType),
		d.PeerIdAddr.String(),
		BleAddrTypeToString(d.PeerOtaAddrType),
		d.PeerOtaAddr.String())
}

type BleEncryptWhen int

const (
	BLE_ENCRYPT_NEVER BleEncryptWhen = iota
	BLE_ENCRYPT_PRIV_ONLY
	BLE_ENCRYPT_ALWAYS
)
