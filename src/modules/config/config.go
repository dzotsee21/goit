package config

import (
	"fmt"
	filesmodule "goit/src/modules/files"
	"regexp"
	"strconv"
	"strings"
)

func IsBare() bool {
	return Read()["core"].(map[string]map[string]string)[""]["bare"] == "true"
}

func ObjectToStr(configObj map[string]interface{}) string {
	objKeys := objectToKeys(configObj)

	obj := make(map[string]interface{})

	for _, section := range objKeys {
		// sectionKeys := objectToKeys(configObj[section].(map[string]interface{}))
		sectionObj := configObj[section].(map[string]interface{})
		subSectionObj := sectionObj

		obj[section] = subSectionObj
	}

	objectStr := ""
	for key, valueObj := range obj {
		if len(key) > 2 && key[0] == '[' && key[len(key)-1] == ']' {
			key = key[1 : len(key)-1]
		}

		section := "[" + key
		settings := ""
		for param, value := range valueObj.(map[string]interface{}) {
			var strVal string
			switch v := value.(type) {
			case string:
				strVal = v
			case bool:
				strVal = strconv.FormatBool(v)
			case map[string]interface{}:
				if param != "" {
					section += " " + "\"" + param + "\""
				}

				m := value.(map[string]interface{})
				for _k, _v := range m {
					param = _k 
					switch sV := _v.(type) {
					case string:
						strVal = sV
					case bool:
						strVal = strconv.FormatBool(sV)
					}
					break
				}
			default:
				strVal = fmt.Sprintf("%v", v)
			}

			settings += fmt.Sprintf("    %s = %s\n", param, strVal)
			section += "]"
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


func Read() map[string]interface{} {
	return StrToObj(filesmodule.Read(filesmodule.GoitPath("config")))
}

func Write(configObj map[string]interface{}) {
	filesmodule.Write(filesmodule.GoitPath("config"), ObjectToStr(configObj))
}

func StrToObj(str string) map[string]interface{} {
	strSplit := strings.Split(str, "\n")
	entryName := ""
	configObj := make(map[string]interface{})

	for _, sptStr := range strSplit {
		re := regexp.MustCompile(`\[(.*?)\]`)

		if re.Match([]byte(sptStr)) {
			entryName = sptStr
			configObj[entryName] = make(map[string]interface{})
		} else {
			if sptStr != "" {
				vals := strings.Split(sptStr, "=")
				param := strings.TrimSpace(vals[0])
				val := strings.TrimSpace(vals[1])
				configObj[entryName].(map[string]interface{})[param] = val
			}
		}
	}

	return configObj
}
