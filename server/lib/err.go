package lib

import (
	"log"
)

func NOP(err error) {}

func NOPCheck(err error) {
	Check(err, NOP)
}

func Check(err error, cb func(error)) {
	if err != nil {
		log.Fatalf(err.Error())
		cb(err)
	}
}
