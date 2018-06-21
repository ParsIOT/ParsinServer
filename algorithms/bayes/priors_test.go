package bayes

import (
	"testing"
	"ParsinServer/glb"
	"ParsinServer/dbm/parameters"
	"ParsinServer/dbm"
)

func TestPriorsThreaded(t *testing.T) {
	assert.Equal(t, OptimizePriorsThreaded("testdb"), nil)
}

// func ExampleTestPriors() {
// 	// optimizePriors("testdb")
// 	fmt.Println("OK")
// 	// Output: OK
// }

//go test -test.bench BenchmarkOptimizePriors -test.benchmem
func BenchmarkOptimizePriors(b *testing.B) {
	for i := 0; i < b.N; i++ {
		optimizePriors("testdb")
	}
}

func BenchmarkOptimizePriorsThreaded(b *testing.B) {
	for i := 0; i < b.N; i++ {
		OptimizePriorsThreaded("testdb")
	}
}

func BenchmarkOptimizePriorsThreadedNot(b *testing.B) {
	for i := 0; i < b.N; i++ {
		OptimizePriorsThreadedNot("testdb")
	}
}

func BenchmarkCrossValidation(b *testing.B) {
	group := "testdb"

	// generate the fingerprintsInMemory
	fingerprintsInMemory := make(map[string]parameters.Fingerprint)
	var fingerprintsOrdering []string
	var err error
	fingerprintsOrdering,fingerprintsInMemory,err = dbm.GetLearnFingerPrints(group,true)
	if err != nil{
		return
	}

	var ps = *parameters.NewFullParameters()
	GetParameters(group, &ps, fingerprintsInMemory, fingerprintsOrdering)
	if glb.RuntimeArgs.GaussianDist {
		calculateGaussianPriors(group, &ps, fingerprintsInMemory, fingerprintsOrdering)
	} else {
		calculatePriors(group, &ps, fingerprintsInMemory, fingerprintsOrdering)
	}
	var results = *parameters.NewResultsParameters()
	for n := range ps.Priors {
		ps.Results[n] = results
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for n := range ps.Priors {
			ps.Priors[n].Special["MixIn"] = 0.5
			ps.Priors[n].Special["VarabilityCutoff"] = 0.005
			crossValidation(group, n, &ps, fingerprintsInMemory, fingerprintsOrdering)
			break
		}
	}
}

func BenchmarkCalculatePriors(b *testing.B) {
	group := "testdb"
	// generate the fingerprintsInMemory
	fingerprintsInMemory := make(map[string]parameters.Fingerprint)
	var fingerprintsOrdering []string
	var err error

	fingerprintsOrdering,fingerprintsInMemory,err = dbm.GetLearnFingerPrints(group,true)
	if err != nil{
		return
	}
	var ps = *parameters.NewFullParameters()
	GetParameters(group, &ps, fingerprintsInMemory, fingerprintsOrdering)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if glb.RuntimeArgs.GaussianDist {
			calculateGaussianPriors(group, &ps, fingerprintsInMemory, fingerprintsOrdering)
		} else {
			calculatePriors(group, &ps, fingerprintsInMemory, fingerprintsOrdering)
		}
	}
}

func BenchmarkGetParameters(b *testing.B) {
	group := "testdb"
	// generate the fingerprintsInMemory
	fingerprintsInMemory := make(map[string]parameters.Fingerprint)
	var fingerprintsOrdering []string
	var err error
	fingerprintsOrdering,fingerprintsInMemory,err = dbm.GetLearnFingerPrints(group,true)
	if err != nil{
		return
	}

	var ps = *parameters.NewFullParameters()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetParameters(group, &ps, fingerprintsInMemory, fingerprintsOrdering)
	}
}
