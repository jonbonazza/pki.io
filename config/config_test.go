package config

import (
	//"fmt"
	//"strings"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestConfigNew(t *testing.T) {
	conf := New("abc")
	assert.NotNil(t, conf)
	assert.Equal(t, conf.Path, "abc")
}

func TestConfigSaveLoad(t *testing.T) {
	file, _ := ioutil.TempFile(".", "xxx")
	filename := file.Name()
	file.Close()

	conf := New(filename)
	conf.AddOrg("org1", "000")
	conf.AddAdmin("admin1", "123")
	conf.AddNode("node1", "456")
	conf.AddNode("node2", "789")
	conf.Save()

	newConf := New(filename)
	err := newConf.Load()
	assert.Nil(t, err)
	assert.Equal(t, conf, newConf)
	os.Remove(filename)
}

func TestConfigLoadNoFile(t *testing.T) {
	conf := New("does_not_exist")
	err := conf.Load()
	assert.NotNil(t, err)
}
