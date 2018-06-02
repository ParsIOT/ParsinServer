package algorithms

import (
	"testing"
	"ParsinServer/glb"
	"github.com/stretchr/testify/assert"
)

func TestGetAccuracyCircleRadius(t *testing.T) {
	dist := GetAccuracyCircleRadius("0,0", []string{"1,1"})
	glb.Debug.Println(dist)
	assert.Equal(t, float64(1.41), dist)
}
