package vici

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

type segmentType byte

const (
	stCMD_REQUEST      segmentType = 0
	stCMD_RESPONSE     segmentType = 1
	stCMD_UNKNOWN      segmentType = 2
	stEVENT_REGISTER   segmentType = 3
	stEVENT_UNREGISTER segmentType = 4
	stEVENT_CONFIRM    segmentType = 5
	stEVENT_UNKNOWN    segmentType = 6
	stEVENT            segmentType = 7
)

func (t segmentType) hasName() bool {
	switch t {
	case stCMD_REQUEST, stEVENT_REGISTER, stEVENT_UNREGISTER, stEVENT:
		return true
	}
	return false
}
func (t segmentType) isValid() bool {
	switch t {
	case stCMD_REQUEST, stCMD_RESPONSE, stCMD_UNKNOWN, stEVENT_REGISTER,
		stEVENT_UNREGISTER, stEVENT_CONFIRM, stEVENT_UNKNOWN, stEVENT:
		return true
	}
	return false
}

func (t segmentType) hasMsg() bool {
	switch t {
	case stCMD_REQUEST, stCMD_RESPONSE, stEVENT:
		return true
	}
	return false
}

type elementType byte

const (
	etSECTION_START elementType = 1
	etSECTION_END   elementType = 2
	etKEY_VALUE     elementType = 3
	etLIST_START    elementType = 4
	etLIST_ITEM     elementType = 5
	etLIST_END      elementType = 6
)

type segment struct {
	typ  segmentType
	name string
	msg  map[string]interface{}
}

// msg can be of three types
// - string
// - map[string]interface{}
// - []string
func writeSegment(w io.Writer, msg segment) error {
	if !msg.typ.isValid() {
		return fmt.Errorf("[writeSegment] msg.typ %d not defined", msg.typ)
	}
	buf := &bytes.Buffer{}
	buf.WriteByte(byte(msg.typ))

	//name
	if msg.typ.hasName() {
		err := writeString1(buf, msg.name)
		if err != nil {
			return fmt.Errorf("write string1 to buffer: %w", err)
		}
	}

	if msg.typ.hasMsg() {
		err := writeMap(buf, msg.msg)
		if err != nil {
			return fmt.Errorf("write map to buffer: %w", err)
		}
	}

	// write length of segment to output
	err := binary.Write(w, binary.BigEndian, uint32(buf.Len()))
	if err != nil {
		return fmt.Errorf("write length: %w", err)
	}

	// write msg to output
	_, err = buf.WriteTo(w)
	if err != nil {
		return fmt.Errorf("write segment: %w", err)
	}

	return nil
}

func readSegment(inR io.Reader) (segment, error) {
	// read length of segment
	var length uint32
	err := binary.Read(inR, binary.BigEndian, &length)
	if err != nil {
		return segment{}, fmt.Errorf("read length: %w", err)
	}
	r := bufio.NewReader(&io.LimitedReader{
		R: inR,
		N: int64(length),
	})
	// read type of segment
	c, err := r.ReadByte()
	if err != nil {
		return segment{}, fmt.Errorf("read type: %w", err)
	}
	var msg segment
	msg.typ = segmentType(c)
	if !msg.typ.isValid() {
		return msg, fmt.Errorf("[readSegment] msg.typ %d not defined", msg.typ)
	}
	if msg.typ.hasName() {
		msg.name, err = readString1(r)
		if err != nil {
			return segment{}, fmt.Errorf("read string1: %w", err)
		}
	}
	if msg.typ.hasMsg() {
		msg.msg, err = readMap(r, true)
		if err != nil {
			return segment{}, fmt.Errorf("read map: %w", err)
		}
	}
	return msg, nil
}

// writeString1 writes string s in a single byte.
func writeString1(w *bytes.Buffer, s string) error {
	length := len(s)
	if length > 255 {
		return fmt.Errorf("[writeString1] string length over 255 characters (1 byte)")
	}
	w.WriteByte(byte(length))
	_, err := w.WriteString(s)
	if err != nil {
		return err
	}
	return nil
}

// readString1 reads one byte from r.
func readString1(r *bufio.Reader) (string, error) {
	length, err := r.ReadByte()
	if err != nil {
		return "", err
	}
	buf := make([]byte, length)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// writeString2 writes string s in a two bytes.
func writeString2(w *bytes.Buffer, s string) error {
	length := len(s)
	if length > 65535 {
		return fmt.Errorf("[writeString2] string length over 65535 characters (2 bytes)")
	}
	err := binary.Write(w, binary.BigEndian, uint16(length))
	if err != nil {
		return err
	}
	_, err = w.WriteString(s)
	if err != nil {
		return err
	}
	return nil
}

// readString2 reads two bytes from r.
func readString2(r io.Reader) (string, error) {
	var length uint16
	err := binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return "", err
	}
	buf := make([]byte, length)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func writeKeyMap(w *bytes.Buffer, name string, msg map[string]interface{}) error {
	w.WriteByte(byte(etSECTION_START))
	err := writeString1(w, name)
	if err != nil {
		return err
	}
	err = writeMap(w, msg)
	if err != nil {
		return err
	}
	w.WriteByte(byte(etSECTION_END))
	return nil
}

func writeKeyList(w *bytes.Buffer, name string, msg []string) error {
	w.WriteByte(byte(etLIST_START))
	err := writeString1(w, name)
	if err != nil {
		return err
	}
	for _, s := range msg {
		w.WriteByte(byte(etLIST_ITEM))
		err = writeString2(w, s)
		if err != nil {
			return err
		}
	}
	w.WriteByte(byte(etLIST_END))
	return nil
}

func writeKeyString(w *bytes.Buffer, name string, msg string) error {
	w.WriteByte(byte(etKEY_VALUE))
	err := writeString1(w, name)
	if err != nil {
		return err
	}
	err = writeString2(w, msg)
	if err != nil {
		return err
	}
	return nil
}

func writeMap(w *bytes.Buffer, msg map[string]interface{}) error {
	for k, v := range msg {
		var err error
		switch t := v.(type) {
		case map[string]interface{}:
			err = writeKeyMap(w, k, t)
		case []string:
			err = writeKeyList(w, k, t)
		case string:
			err = writeKeyString(w, k, t)
		case []interface{}:
			str := make([]string, len(t))
			for i := range t {
				str[i] = t[i].(string)
			}
			err = writeKeyList(w, k, str)
		default:
			return fmt.Errorf("[writeMap] can not write type %T right now", msg)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

//SECTION_START has been read already.
func readKeyMap(r *bufio.Reader) (string, map[string]interface{}, error) {
	key, err := readString1(r)
	if err != nil {
		return "", nil, err
	}
	msg, err := readMap(r, false)
	if err != nil {
		return "", nil, err
	}
	return key, msg, nil
}

//LIST_START has been read already.
func readKeyList(r *bufio.Reader) (string, []string, error) {
	key, err := readString1(r)
	if err != nil {
		return "", nil, err
	}
	msg := []string{}
	for {
		var c byte
		c, err = r.ReadByte()
		if err != nil {
			return "", nil, err
		}
		switch elementType(c) {
		case etLIST_ITEM:
			value, err := readString2(r)
			if err != nil {
				return "", nil, err
			}
			msg = append(msg, value)
		case etLIST_END: //end of outer list
			return key, msg, nil
		default:
			return "", nil, fmt.Errorf("[readKeyList] protocol error 2")
		}
	}
}

//KEY_VALUE has been read already.
func readKeyString(r *bufio.Reader) (string, string, error) {
	key, err := readString1(r)
	if err != nil {
		return "", "", err
	}
	msg, err := readString2(r)
	if err != nil {
		return "", "", err
	}
	return key, msg, nil
}

// Since the original key chosen can have duplicates,
// this function is used to map the original key to a new one
// to make them unique.
func getNewKeyToHandleDuplicates(key string, msg map[string]interface{}) string {
	if _, ok := msg[key]; !ok {
		return key
	}

	for i := 0; ; i++ {
		newKey := key + "##" + strconv.Itoa(i)
		if _, ok := msg[newKey]; !ok {
			return newKey
		}
	}
}

//SECTION_START has been read already.
func readMap(r *bufio.Reader, isRoot bool) (map[string]interface{}, error) {
	msg := map[string]interface{}{}
	for {
		c, err := r.ReadByte()
		if err == io.EOF && isRoot { //may be root section
			return msg, nil
		}
		if err != nil {
			return nil, err
		}
		switch elementType(c) {
		case etSECTION_START:
			key, value, err := readKeyMap(r)
			if err != nil {
				return nil, err
			}
			msg[getNewKeyToHandleDuplicates(key, msg)] = value
		case etLIST_START:
			key, value, err := readKeyList(r)
			if err != nil {
				return nil, err
			}
			msg[getNewKeyToHandleDuplicates(key, msg)] = value
		case etKEY_VALUE:
			key, value, err := readKeyString(r)
			if err != nil {
				return nil, err
			}
			msg[getNewKeyToHandleDuplicates(key, msg)] = value
		case etSECTION_END: //end of outer section
			return msg, nil
		default:
			return nil, fmt.Errorf("[readMap] protocol error 1, %d", c)
		}
	}
}
