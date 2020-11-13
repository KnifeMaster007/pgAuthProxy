package utils

import log "github.com/sirupsen/logrus"

type CustomLoggerHolder struct {
	loggerProperties map[string]interface{}
}

func NewLoggerHolder(props map[string]interface{}) *CustomLoggerHolder {
	return &CustomLoggerHolder{loggerProperties: props}
}

func (lh *CustomLoggerHolder) SetProperty(key string, value string) {
	lh.loggerProperties[key] = value
}

func (lh *CustomLoggerHolder) Get() *log.Entry {
	return log.WithFields(lh.loggerProperties)
}
