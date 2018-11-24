// +build public
// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

func (d *DiagnosticData) decodeFTDC(data []byte) error {
	d.unmarshalFirstBsonDoc(data)
	return nil
}
