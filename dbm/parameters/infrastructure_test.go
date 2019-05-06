package parameters

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddChangeTransmitters(t *testing.T) {
	t1 := Transmitter{Mac: "mac1", Location: "1,1"}
	t2 := Transmitter{Mac: "mac2", Location: "2,2"}
	t3 := Transmitter{Mac: "mac3", Location: "3,3"}

	inf := Infrastructure{
		Transmitters: []Transmitter{t1, t2, t3},
	}

	t1Changed := Transmitter{Mac: "mac1", Location: "-1,-1"}
	t4 := Transmitter{Mac: "mac4", Location: "4,4"}

	resInf := AddChangeTransmitters(inf, []Transmitter{t1Changed, t4})

	expectedResInf := Infrastructure{
		Transmitters: []Transmitter{t1Changed, t2, t3, t4},
	}

	assert.Equal(t, expectedResInf.Transmitters, resInf.Transmitters)
}

func TestDelTransmitters(t *testing.T) {
	t1 := Transmitter{Mac: "mac1", Location: "1,1"}
	t2 := Transmitter{Mac: "mac2", Location: "2,2"}
	t3 := Transmitter{Mac: "mac3", Location: "3,3"}

	inf := Infrastructure{
		Transmitters: []Transmitter{t1, t2, t3},
	}

	t4NotExists := Transmitter{Mac: "mac4", Location: "4,4"}

	resInf := DelTransmitters(inf, []Transmitter{t1, t4NotExists})

	expectedResInf := Infrastructure{
		Transmitters: []Transmitter{t2, t3},
	}

	assert.Equal(t, expectedResInf.Transmitters, resInf.Transmitters)
}
