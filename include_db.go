// Copyright Kuei-chun Chen, 2021-present. All rights reserved.

package keyhole

// IncludeDB stores excluded QueryFilter
type IncludeDB []string

func (p *IncludeDB) String() string {
	str := ""
	for _, ns := range *p {
		str += "-db " + ns + " "
	}
	return str
}

// Set sets flag var
func (p *IncludeDB) Set(value string) error {
	*p = append(*p, value)
	return nil
}
