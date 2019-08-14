package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/blend/go-sdk/assert"
)

func TestNewErrorEvent(t *testing.T) {
	assert := assert.New(t)

	/// stuff
	ee := NewErrorEvent(
		Fatal,
		fmt.Errorf("not a test"),
	)
	assert.Equal(Fatal, ee.GetFlag())
	assert.Equal("not a test", ee.Err.Error())

	buf := new(bytes.Buffer)
	tf := TextOutputFormatter{
		NoColor: true,
	}

	ee.WriteText(tf, buf)
	assert.Equal("not a test", buf.String())

	contents, err := json.Marshal(ee.Decompose())
	assert.Nil(err)
	assert.Contains(string(contents), "not a test")
}

func TestErrorEventListener(t *testing.T) {
	assert := assert.New(t)

	ee := NewErrorEvent(Fatal, fmt.Errorf("only a test"))

	var didCall bool
	ml := NewErrorEventListener(func(ctx context.Context, e ErrorEvent) {
		didCall = true
	})

	ml(context.Background(), ee)
	assert.True(didCall)
}
