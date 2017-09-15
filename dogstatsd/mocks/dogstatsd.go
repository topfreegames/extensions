// maestro
// https://github.com/topfreegames/maestro
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package mocks

type DogStatsDClientMock struct {
	Counts map[string]int64
}

func NewDogStatsDClientMock() *DogStatsDClientMock {
	return &DogStatsDClientMock{
		Counts: map[string]int64{},
	}
}

func (d *DogStatsDClientMock) Increment(name string, tags []string, rate float64) error {
	d.Counts[name]++
	return nil
}
