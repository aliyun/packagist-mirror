package util

type Mirror struct {
	providerUrl          string
	distUrl              string
	apiIterationInterval int
}

func NewMirror(providerUrl string, distUrl string, apiIterationInterval int) (mirror *Mirror) {
	return &Mirror{
		providerUrl:          providerUrl,
		distUrl:              distUrl,
		apiIterationInterval: apiIterationInterval,
	}
}
