package controllers

import (
	"encoding/json"
	"strconv"
)

type MvcController struct{}

func (m *MvcController) Get() (string, error) {

	someMap := map[int]int{}

	for i := 1; i < 10; i++ {
		someMap[i] = i
	}

	b, err := json.Marshal(someMap)

	return string(b), err
}

func (m *MvcController) Post() string {

	res := 1
	for i := 1; i < 10; i++ {
		res *= i
	}

	return strconv.Itoa(res)
}
