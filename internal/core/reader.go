package core

import (
	"encoding/binary"
	"errors"
	"io"
)

// RecordReader reads records from a stream.
type RecordReader struct {
	reader io.Reader
	stash  []byte
}

// NewRecordReader creates a new record reader.
func NewRecordReader(reader io.Reader) *RecordReader {
	return &RecordReader{reader: reader}
}

// Read implements io.Reader, reassembling records into continuous data.
func (r *RecordReader) Read(p []byte) (int, error) {
	for len(r.stash) == 0 {
		record, err := r.ReadNextRecord()
		if err != nil {
			return 0, err
		}
		if record.Type == TypeError {
			return 0, errors.New("server error: " + record.ErrorMessage)
		}
		if record.Type != TypeData {
			continue
		}
		r.stash = record.Payload
	}

	n := copy(p, r.stash)
	r.stash = r.stash[n:]
	return n, nil
}

// ReadNextRecord reads and parses a single record.
func (r *RecordReader) ReadNextRecord() (*Record, error) {
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(r.reader, lengthBytes); err != nil {
		return nil, err
	}
	
	totalLength := binary.BigEndian.Uint32(lengthBytes)
	if totalLength < RecordHeaderLength {
		return nil, errors.New("invalid record length")
	}

	recordBytes := make([]byte, totalLength)
	if _, err := io.ReadFull(r.reader, recordBytes); err != nil {
		return nil, err
	}

	recordType := recordBytes[0]
	payloadLength := binary.BigEndian.Uint32(recordBytes[4:8])
	paddingLength := binary.BigEndian.Uint32(recordBytes[8:12])
	iv := recordBytes[12:24]

	if int(RecordHeaderLength+payloadLength+paddingLength) != len(recordBytes) {
		return nil, errors.New("invalid payload length")
	}

	header := recordBytes[:RecordHeaderLength]
	payloadStart := RecordHeaderLength
	payloadEnd := payloadStart + int(payloadLength)
	payload := recordBytes[payloadStart:payloadEnd]

	result := &Record{Type: recordType, Payload: payload, Header: header, IV: iv}
	if recordType == TypeError {
		if len(payload) >= 4 {
			// Skip 2 bytes error code?
			// Worker.js: writeUint16(payload, 0, code); payload.set(messageBytes, 4);
			// So payload[0-1] is code. payload[4:] is message?
			// Wait, worker.js:
			// const payload = new Uint8Array(4 + messageBytes.byteLength);
			// writeUint16(payload, 0, code);
			// payload.set(messageBytes, 4);
			// So bytes 2 and 3 are 0?
			// Yes.
			result.ErrorMessage = string(payload[4:])
		}
	}
	return result, nil
}
