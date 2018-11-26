// +build !delta
// Copyright 2018 Kuei-chun Chen. All rights reserved.

package keyhole

func getDiagnosticData(data []byte, span int) (DiagnosticData, error) {
	return unmarshalFirstBsonDoc(data), nil
}
