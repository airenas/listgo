package utils

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLJoin(t *testing.T) {
	assert.Equal(t, "http://www.delfi.lt/olia", URLJoin("http://www.delfi.lt", "olia"))
	assert.Equal(t, "http://www.delfi.lt/olia/1", URLJoin("http://www.delfi.lt", "olia", "1"))
	assert.Equal(t, "http://www.delfi.lt/olia/1", URLJoin("http://www.delfi.lt/", "/olia/", "1"))
	assert.Equal(t, "http://www.delfi.lt/olia/1", URLJoin("http://www.delfi.lt", "olia", "/1"))
	assert.Equal(t, "http://www.delfi.lt", URLJoin("http://www.delfi.lt"))
	assert.Equal(t, "http://www.delfi.lt:80/olia", URLJoin("http://www.delfi.lt:80/", "olia"))
	assert.Equal(t, "www.delfi.lt:80/olia", URLJoin("www.delfi.lt:80", "olia"))
}

func TestValidateURL(t *testing.T) {
	ut, err := validateConfigURL("http://www.delfi.lt/olia/1", "sn")
	assert.Equal(t, "http://www.delfi.lt/olia/1", ut)
	assert.Nil(t, err)
}

func TestValidateURL_FailEmpty(t *testing.T) {
	ut, err := validateConfigURL("", "sn")
	assert.Equal(t, "", ut)
	assert.NotNil(t, err)
}

func TestValidateURL_Fail(t *testing.T) {
	ut, err := validateConfigURL(":::://", "sn")
	assert.Equal(t, "", ut)
	assert.NotNil(t, err)
}

func TestValidateResponse(t *testing.T) {
	r := http.Response{StatusCode: 200}
	err := ValidateResponse(&r)
	assert.Nil(t, err)
}

func TestValidateResponseBadParam(t *testing.T) {
	r := http.Response{StatusCode: 400, Body: ioutil.NopCloser(bytes.NewReader([]byte("errorX")))}
	err := ValidateResponse(&r)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, ErrWrongHTTPCall))
}

func TestValidateResponseNotBadParam(t *testing.T) {
	r := http.Response{StatusCode: 500, Body: ioutil.NopCloser(bytes.NewReader([]byte("errorX")))}
	err := ValidateResponse(&r)
	assert.NotNil(t, err)
	assert.False(t, errors.Is(err, ErrWrongHTTPCall))
}

func TestValidateResponseTakesBody(t *testing.T) {
	r := http.Response{StatusCode: 400, Body: ioutil.NopCloser(bytes.NewReader([]byte("errorX")))}
	err := ValidateResponse(&r)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "errorX"))
}
