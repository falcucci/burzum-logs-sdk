# Logzum - logrus Burzumlogs Hook
Logzum is a logrus hook for BurzumLogs



## Install
```
go get github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/logzum
```


## Usage
```
	// Initialize the hook with default configs
	hook, err := logzum.New("bztoken")

	if err != nil {
		log.Printf("Could not initialize Logzum: %s", err.Error())
		return
	}

	//add the hook to the logrus
	logrus.AddHook(hook)
```
If you want to change the log level, you can change the configuration of the hook:
```
	// Creates default settings
	config := logzum.DefaultConfig
	// Change the log level
	config.MinLevel = logrus.InfoLevel
	// Initialize the hook with your configs
	hook, err := logzum.NewWithConfig("bztoken", config)

	if err != nil {
		log.Printf("Could not initialize Logzum: %s", err.Error())
		return
	}

	//add the hook to the logrus
	logrus.AddHook(hook)
```