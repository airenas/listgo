package utils

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/pkg/errors"
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

func TestHidePass(t *testing.T) {
	assert.Equal(t, "", HidePass(""))
	assert.Equal(t, "http://delfi.lt", HidePass("http://delfi.lt"))
	assert.Equal(t, "http://delfi.lt:8080/aaa", HidePass("http://delfi.lt:8080/aaa"))
	assert.Equal(t, "http://user:----@delfi.lt:8080/aaa", HidePass("http://user:olia@delfi.lt:8080/aaa"))
	assert.Equal(t, "http://user:----@delfi.lt/aaa", HidePass("http://user:olia@delfi.lt/aaa"))
	assert.Equal(t, "ampq://user:----@delfi.lt/aaa", HidePass("ampq://user:olia@delfi.lt/aaa"))
}

func TestHidePass_Mongo(t *testing.T) {
	assert.Equal(t, "mongodb://mongo:27017", HidePass("mongodb://mongo:27017"))
	assert.Equal(t, "mongodb://l:----@mongo:27017", HidePass("mongodb://l:olia@mongo:27017"))
}

func TestSupportAudioExt(t *testing.T) {
	assert.True(t, SupportAudioExt(".wav"))
	assert.True(t, SupportAudioExt(".mp4"))
	assert.True(t, SupportAudioExt(".mp3"))
	assert.True(t, SupportAudioExt(".m4a"))

	assert.False(t, SupportAudioExt(""))
	assert.False(t, SupportAudioExt(".txt"))
	assert.False(t, SupportAudioExt(".mpeg"))
}

func TestParamTrue(t *testing.T) {
	tests := []struct {
		name string
		args string
		want bool
	}{
		{name: "true", args: "true", want: true},
		{name: "True", args: "True", want: true},
		{name: "TRUE", args: "TRUE", want: true},
		{name: "1", args: "1", want: true},
		{name: "False", args: "false", want: false},
		{name: "False", args: "0", want: false},
		{name: "False", args: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParamTrue(tt.args); got != tt.want {
				t.Errorf("ParamTrue() = %v, want %v", got, tt.want)
			}
		})
	}
}
