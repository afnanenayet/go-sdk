package configutil

import (
	"testing"
	"time"

	"github.com/blend/go-sdk/assert"
)

func TestParse(t *testing.T) {
	assert := assert.New(t)

	stringSource := String("")
	intValue, err := Parse(stringSource).Int()
	assert.Nil(err)
	assert.Nil(intValue)

	stringSource = String("bad value")
	intValue, err = Parse(stringSource).Int()
	assert.NotNil(err)
	assert.Nil(intValue)

	stringSource = String("1234")
	intValue, err = Parse(stringSource).Int()
	assert.Nil(err)
	assert.NotNil(intValue)
	assert.Equal(1234, *intValue)

	stringSource = String("")
	floatValue, err := Parse(stringSource).Float64()
	assert.Nil(err)
	assert.Nil(floatValue)

	stringSource = String("bad value")
	floatValue, err = Parse(stringSource).Float64()
	assert.NotNil(err)
	assert.Nil(floatValue)

	stringSource = String("1234.34")
	floatValue, err = Parse(stringSource).Float64()
	assert.Nil(err)
	assert.NotNil(floatValue)
	assert.Equal(1234.34, *floatValue)

	stringSource = String("")
	durationValue, err := Parse(stringSource).Duration()
	assert.Nil(err)
	assert.Nil(durationValue)

	stringSource = String("bad value")
	durationValue, err = Parse(stringSource).Duration()
	assert.NotNil(err)
	assert.Nil(durationValue)

	stringSource = String("10s")
	durationValue, err = Parse(stringSource).Duration()
	assert.Nil(err)
	assert.NotNil(durationValue)
	assert.Equal(10*time.Second, *durationValue)
}
