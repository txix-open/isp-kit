package metrics

var (
	DefaultObjectives = map[float64]float64{
		0.5:  0.05,  //0.45-0.55
		0.9:  0.01,  //0.89-0.91
		0.95: 0.005, //0.94.5-0.95.5
		0.99: 0.001, //0.98.9-0.99.1,
	}
)
