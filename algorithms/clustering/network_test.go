package clustring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildNetwork(t *testing.T) {
	newNetwork := make(map[string]map[string]bool)
	newNetwork = BuildNetwork(newNetwork, []string{"pie", "ice cream"})
	newNetwork = BuildNetwork(newNetwork, []string{"pie", "cocolate syrup"})
	newNetwork = BuildNetwork(newNetwork, []string{"orange juice", "water"})
	newNetwork = BuildNetwork(newNetwork, []string{"water", "coffee"})
	assert.Equal(t, newNetwork["0"], map[string]bool{"ice cream": true, "cocolate syrup": true, "pie": true})
}

func TestHasNetwork(t *testing.T) {
	newNetwork := make(map[string]map[string]bool)
	newNetwork = BuildNetwork(newNetwork, []string{"pie", "ice cream"})
	newNetwork = BuildNetwork(newNetwork, []string{"pie", "cocolate syrup"})
	newNetwork = BuildNetwork(newNetwork, []string{"orange juice", "water"})
	newNetwork = BuildNetwork(newNetwork, []string{"water", "coffee"})
	network, _ := HasNetwork(newNetwork, []string{"water"})
	assert.Equal(t, network, "1")
}
