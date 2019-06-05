package file

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"github.com/stretchr/testify/assert"
)

func initFiler(t *testing.T) (string, *Filer) {
	dir, err := ioutil.TempDir("", "prefix")
	assert.Nil(t, err)
	f := Filer{path: dir}
	return dir, &f
}

func TestFind_Empty(t *testing.T) {
	d, f := initFiler(t)
	defer os.RemoveAll(d)

	m, err := f.Find("olia")
	assert.Nil(t, err)
	assert.Nil(t, m)
}

func TestFind_Exists(t *testing.T) {
	d, f := initFiler(t)
	defer os.RemoveAll(d)
	ioutil.WriteFile(path.Join(d, "olia"), []byte("id1"), 0644)

	m, err := f.Find("olia")
	assert.Nil(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, m.TrID, "id1")
	assert.Equal(t, m.KafkaID, "olia")
}

func TestSetWorking(t *testing.T) {
	d, f := initFiler(t)
	defer os.RemoveAll(d)
	m, err := f.Find("olia")
	assert.Nil(t, err)
	assert.Nil(t, m)

	err = f.SetWorking(&kafkaapi.KafkaTrMap{KafkaID: "olia", TrID: "1"})

	m, err = f.Find("olia")
	assert.Nil(t, err)
	assert.NotNil(t, m)
}

func TestDelete(t *testing.T) {
	d, f := initFiler(t)
	defer os.RemoveAll(d)
	ioutil.WriteFile(path.Join(d, "olia"), []byte("id1"), 0644)

	m, err := f.Find("olia")
	assert.Nil(t, err)
	assert.NotNil(t, m)

	err = f.Delete("olia")

	m, err = f.Find("olia")
	assert.Nil(t, err)
	assert.Nil(t, m)
}
