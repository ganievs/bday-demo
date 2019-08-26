package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

var (
	invalidFromat = "2006.01.02"
	invalidDate   = "2080-01-02"
	invalidUser   =  "User123"
)

func TestOnlyLetters(t *testing.T) {
	u := onlyLetters(invalidUser)
	assert.False(t, u)
}

func TestCheckinvalidFormat(t *testing.T) {
	d := checkDate(invalidFromat)
	assert.False(t, d)
}

func TestCheckInvalidDate(t *testing.T) {
	d := checkDate(invalidDate)
	assert.False(t, d)
}
