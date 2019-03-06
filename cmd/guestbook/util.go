package main

import (
	log "github.com/sirupsen/logrus"
)

// logIfError logs errors when they happen
func logIfError(result interface{}, err error) (r interface{}) {
	if err != nil {
		log.Error(err)
	}
	return result
}
