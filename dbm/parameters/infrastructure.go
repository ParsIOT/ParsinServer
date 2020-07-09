package parameters

// ItemGraph the Items graph

type Transmitter struct {
	Mac      string `json:"Mac"`
	Location string `json:"Location"`
}

type Infrastructure struct {
	Transmitters []Transmitter `json:"Transmitters"` //mac --> locations
}

func NewInfrastructure() Infrastructure {
	return Infrastructure{
		Transmitters: []Transmitter{},
	}
}

// Change a transmitter if it exists in inf.Transmitters or append it if doesn't
func AddChangeTransmitters(inf Infrastructure, transmitters []Transmitter) Infrastructure {
	resInf := NewInfrastructure()
	changesTransmitters := []Transmitter{}

	// Change existent transmitters
	for _, oldTrans := range inf.Transmitters {
		foundTrans := false
		for _, newTrans := range transmitters {
			if newTrans.Mac == oldTrans.Mac {
				resInf.Transmitters = append(resInf.Transmitters, newTrans)
				changesTransmitters = append(changesTransmitters, newTrans)
				foundTrans = true
				break
			}
		}
		if !foundTrans { // append as new new transmitter
			resInf.Transmitters = append(resInf.Transmitters, oldTrans)
		}
	}

	// Add new transmitters
	for _, newTrans := range transmitters {
		foundTrans := false
		for _, changedTrans := range changesTransmitters {
			if newTrans.Mac == changedTrans.Mac {
				foundTrans = true
			}
		}
		if !foundTrans { // append as new new transmitter
			resInf.Transmitters = append(resInf.Transmitters, newTrans)
		}
	}
	return resInf
}

func DelTransmitters(inf Infrastructure, transmitter []Transmitter) Infrastructure {
	resInf := NewInfrastructure()
	for _, trans := range inf.Transmitters {
		transFound := false
		for _, delTrans := range transmitter {
			if trans.Mac == delTrans.Mac {
				transFound = true
				break
			}
		}
		if !transFound {
			resInf.Transmitters = append(resInf.Transmitters, trans)
		}
	}
	return resInf

}
