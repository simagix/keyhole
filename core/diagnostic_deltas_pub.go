// +build public
// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

func decodeFTDC(data []byte, span int) (DiagnosticData, error) {
	return unmarshalFirstBsonDoc(data), nil
}
