// Copyright 2018 Kuei-chun Chen. All rights reserved.

package mdb

// GetKeys gets all fields of a odc as an array
func GetKeys(filter map[string]interface{}) []string {
	var arr []string
	for key, val := range filter {
		if key == "$or" || key == "$and" {
			for _, elem := range val.([]interface{}) {
				for k := range elem.(map[string]interface{}) {
					if isKeyword(k) == false {
						arr = append(arr, k)
					}
				}
			}
		} else if isKeyword(key) == false {
			arr = append(arr, key)
		}
	}
	return arr
}

func isKeyword(key string) bool {
	return len(key) > 0 && key[0] == '$'
}
