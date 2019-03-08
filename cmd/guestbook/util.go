package main

import (
	"math"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

// logIfError logs errors when they happen
func logIfError(result interface{}, err error) (r interface{}) {
	if err != nil {
		log.Error(err)
	}
	return result
}

// chance of something happening
func chance(percent float64) bool {
	c := (float64(rand.Intn(1000)) / 100.0)

	return (c <= math.Abs(percent))
}

// randomSleep sometime between zero and max milliseconds
func randomSleep(max uint) {
	sleepTime := rand.Intn(int(max))
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
}
