package cmdworker

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun_NoParameter_Fail(t *testing.T) {
	cmd := "ls"
	err := RunCommand(cmd, "/", "id", nil, ioutil.Discard)

	assert.NotNil(t, err, "Error expected")
}

func TestRun_WrongParameter_Fail(t *testing.T) {
	cmd := "ls -{olia}"
	err := RunCommand(cmd, "/", "id", nil, ioutil.Discard)

	assert.NotNil(t, err, "Error expected")
}
func TestRun(t *testing.T) {
	cmd := "ls -la"
	err := RunCommand(cmd, "/", "id", nil, ioutil.Discard)
	assert.Nil(t, err)
}

func TestRun_ID_Changed(t *testing.T) {
	cmd := "ls -{ID}"
	err := RunCommand(cmd, "/", "la", nil, ioutil.Discard)
	assert.Nil(t, err)
}

func TestRun_WritesToLog(t *testing.T) {
	cmd := "echo olia"
	var b bytes.Buffer
	err := RunCommand(cmd, "/", "id", nil, &b)
	s := string(b.Bytes())
	assert.Nil(t, err, "Error not expected")
	assert.Contains(t, s, "\nolia\n")
	assert.Contains(t, s, "Running")
	assert.Contains(t, s, "Finished")
}

func TestRun_WritesOnError(t *testing.T) {
	cmd := "ech olia"
	var b bytes.Buffer
	err := RunCommand(cmd, "/", "id", nil, &b)
	s := string(b.Bytes())
	assert.NotNil(t, err, "Error expected")
	assert.Contains(t, s, "Running")
	assert.Contains(t, s, "ERROR")
	assert.Contains(t, err.Error(), "ech")
}
