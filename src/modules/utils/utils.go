package utils

import (
	"log"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
)

func Lines(str string) []string {
	splitStr := strings.Split(str, "\n")

	var lines []string
	for _, val := range splitStr {
		if val != "" {
			lines = append(lines, val)
		}
	}

	return lines
}

func SetIn(obj map[string]interface{}, arr []interface{}) map[string]interface{} {
	if len(arr) == 2 {
		obj[arr[0].(string)] = arr[1]
	}

	if len(arr) > 2 {
		_, exists := obj[arr[0].(string)]
		if !exists {
			obj[arr[0].(string)] = map[string]interface{}{}
		}

		SetIn(obj[arr[0].(string)].(map[string]interface{}), arr[1:])
	}

	return obj
}

func MapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func Hash(str string) string {
	hashInt := 0
	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		code := int(runes[i])
		hashInt = hashInt*31 + code
	}
	return strconv.FormatFloat(math.Abs(float64(hashInt)), 'f', -1, 64)
}

func Intersection(a, b []string) []string {
	var result []string
	for _, e := range a {
		if slices.Contains(b, e) {
			result = append(result, e)
		}
	}

	return result
}

func IsDir(path string) bool {
	info, _ := os.Stat(path)

	return info.IsDir()
}

func Unique(arr []string) []string {
	var uElements []string
	for _, key := range arr {
		if !slices.Contains(uElements, key) {
			uElements = append(uElements, key)
		}
	}

	return uElements
}

func OnRemote(remotePath string) func(fn func(interface{}) interface{}, arg ...string) interface{} {
	return func(fn func(interface{}) interface{}, arg ...string) interface{} {
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		err := os.Chdir(remotePath)
		if err != nil {
			log.Fatal(err)
		}

		return fn(arg)
	}
}

func Flatten(arr []interface{}) []interface{} {
	var result []interface{}
	for _, e := range arr {
		if nested, ok := e.([]interface{}); ok {
			result = append(result, Flatten(nested)...)
		} else {
			result = append(result, e)
		}
	}
	return result
}
