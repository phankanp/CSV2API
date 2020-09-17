package helper

import (
	"github.com/phankanp/csv-to-json/model"
	"strconv"
)

func IntFromString(v string) (int, error) {
	res, err := strconv.Atoi(v)

	if err != nil {
		return 0, err
	}

	return res, err
}

func StringInSlice(newHeader string, list []model.Header) bool {
	for _, oldHeader := range list {
		if oldHeader.Name == newHeader {
			return true
		}
	}
	return false
}

func CompareHeaders(newHeaders model.JSONB, oldHeaders []model.Header) bool {
	for key, _ := range newHeaders {
		ok := StringInSlice(key, oldHeaders)

		if !ok {
			return false
		}
	}
	return true
}
