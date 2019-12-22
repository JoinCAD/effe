package effe

import (
	"fmt"
	"strconv"
)

type float64NumberProvider struct{}

type float64Number float64

func (f float64Number) String() string {
	return fmt.Sprint(float64(f))
}

func (p float64NumberProvider) ParseNumber(text string) (Number, error) {
	f, err := strconv.ParseFloat(text, 64)
	return float64Number(f), err
}

func (p float64NumberProvider) Add(a Number, b Number) (Number, error) {
	af := a.(float64Number)
	bf := b.(float64Number)
	return af + bf, nil
}
