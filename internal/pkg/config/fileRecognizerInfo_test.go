package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Unmarshal_Empty(t *testing.T) {
	r, err := loadYaml([]byte(""))
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func Test_Unmarshal_Fails(t *testing.T) {
	r, err := loadYaml([]byte("name:olia\n"))
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func Test_Unmarshal_Name(t *testing.T) {
	r, err := loadYaml([]byte("name: olia"))
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "olia", r.Name)
}

func Test_Unmarshal_Description(t *testing.T) {
	r, err := loadYaml([]byte("name: olia\ndescription: olia"))
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "olia", r.Description)
}

func Test_Unmarshal_DateCreated(t *testing.T) {
	r, err := loadYaml([]byte("name: olia\ndate_created: 2019-11-23"))
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "2019-11-23", r.DateCreated.Format("2006-01-02"))
}

func Test_Unmarshal_Settings(t *testing.T) {
	r, err := loadYaml([]byte("name: olia\nsettings:\n  Model_path: /models/path\n  punctuate: yes"))
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.NotNil(t, r.Settings)
	v, _ := r.Settings["Model_path"]
	assert.Equal(t, "/models/path", v)
	v, _ = r.Settings["punctuate"]
	assert.Equal(t, "yes", v)
}

func Test_LoadFromFile(t *testing.T) {
	f := createTempFile(t)
	defer os.Remove(f.Name())
	fmt.Fprint(f, "name: olia")
	r, err := loadFile(f.Name())
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "olia", r.Name)
}

func Test_LoadFromFile_Fails(t *testing.T) {
	r, err := loadFile("some non existing file")
	assert.NotNil(t, err)
	assert.Nil(t, r)
}
