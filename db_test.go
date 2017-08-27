package main

import (
	"testing"
	"encoding/json"
	"github.com/stretchr/testify/assert"
)

func TestgetAdminUsers(t *testing.T) {
	adminList, err := getAdminUsers("test")
	if err != nil {
		t.Errorf("Can't Unmarshal admin list")
		return
	}
	adminListJson, err := json.Marshal(adminList)
	if err != nil {
		t.Errorf("Can't remarshal admin list!")
	}
	response := `{"admin":"admin"}`
	assert.Equal(t, adminListJson, response)
}
