package logzum_test

import (
	"log"

	"github.com/sirupsen/logrus"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/logzum"
)

func ExampleNew() {
	// Initialize the hook with default configs
	hook, err := logzum.New("bztoken")

	if err != nil {
		log.Printf("Could not initialize Logzum: %s", err.Error())
		return
	}

	//add the hook to the logrus
	logrus.AddHook(hook)

}
