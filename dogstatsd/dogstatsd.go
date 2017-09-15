package dogstatsd

var Value = true

type DogStatsDClient interface {
	Increment(name string, tags []string, rate float64) error
}

type DogStatsD struct {
	client DogStatsDClient
}

func NewDogStatsD(client DogStatsDClient) *DogStatsD {
	return &DogStatsD{
		client: client,
	}
}

func (d *DogStatsD) Increment(name string, tags []string, rate float64) error {
	return d.client.Increment(name, tags, rate)
}
