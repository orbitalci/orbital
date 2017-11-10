package nsqpb

import (
	"github.com/shankj3/ocelot/util/ocelog"
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
	"strings"
)

func ConvertLogLevel(level logrus.Level) nsq.LogLevel {
	switch level {
	case logrus.DebugLevel:
		return nsq.LogLevelDebug
	case logrus.InfoLevel:
		return nsq.LogLevelInfo
	case logrus.WarnLevel:
		return nsq.LogLevelWarning
	case logrus.ErrorLevel:
		return nsq.LogLevelError
	}
	return nsq.LogLevelWarning
}

var (
	nsqDebugLevel = nsq.LogLevelDebug.String()
	nsqInfoLevel  = nsq.LogLevelInfo.String()
	nsqWarnLevel  = nsq.LogLevelWarning.String()
	nsqErrLevel   = nsq.LogLevelError.String()
)

type NSQLogger struct{}

func NewNSQLogger() (logger NSQLogger, level nsq.LogLevel) {
	return NewNSQLoggerAtLevel(ocelog.GetLogLevel())
}

func NewNSQLoggerAtLevel(lvl logrus.Level) (logger NSQLogger, level nsq.LogLevel){
	logger = NSQLogger{}
	level = ConvertLogLevel(lvl)
	return
}

func (n NSQLogger) Output(_ int, s string) error {
	if len(s) > 3 {
		msg := strings.TrimSpace(s[3:])
		switch s[:3] {
		case nsqDebugLevel:
			ocelog.Log().Debugln(msg)
		case nsqInfoLevel:
			ocelog.Log().Infoln(msg)
		case nsqWarnLevel:
			ocelog.Log().Warnln(msg)
		case nsqErrLevel:
			ocelog.Log().Errorln(msg)
		default:
			ocelog.Log().Infoln(msg)
		}
	}
	return nil
}
