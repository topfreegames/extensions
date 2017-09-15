// maestro
// https://github.com/topfreegames/maestro
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

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
