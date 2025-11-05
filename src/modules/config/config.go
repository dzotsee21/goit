package config

import (
	"fmt"
	"strconv"
)

func ObjectToStr(configObj map[string]interface{}) string {
	objKeys := objectToKeys(configObj)

	obj := make(map[string]interface{})

	for _, section := range objKeys {
		sectionKeys := objectToKeys(configObj[section].(map[string]interface{}))
		for _, sk := range sectionKeys {
			sectionObj := configObj[section].(map[string]interface{})
			subSectionObj := sectionObj[sk]

			obj[section] = subSectionObj
		}
	}

	objectStr := ""
	for key, valueObj := range obj {
		section := "[" + key + "]"
		settings := ""
		for param, value := range valueObj.(map[string]interface{}) {
			var strVal string
			switch v := value.(type) {
			case string:
				strVal = v
			case bool:
				strVal = strconv.FormatBool(v)
			default:
				strVal = fmt.Sprintf("%v", v)
			}

			settings += fmt.Sprintf("    %s = %s\n", param, strVal)
		}

		objectStr += section + "\n"
		objectStr += settings + "\n"
	}

	return objectStr

}

func objectToKeys(configObj map[string]interface{}) []string {
	objKeys := make([]string, 0, len(configObj))

	for k := range configObj {
		objKeys = append(objKeys, k)
	}

	return objKeys
}
