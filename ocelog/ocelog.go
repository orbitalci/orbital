/* Package ocelog
Way to have one style of logging for the project. Initialize in your service w/ InitializeOcelog(), uses a JSONFormatter.
use ocelog.Log to log with extra field of the function called.

todo: add common log functions, right now there is only LogErrField which adds the error: <error text> to the json log.

*/
package ocelog

import (
    "flag"
    "os"
    "runtime"
    "strings"
	log "github.com/sirupsen/logrus"
)

var DefaultFields = log.Fields{
	"function": getCaller(),
}

// tODO: add NewLog() function that returns a logger w/ the default fields, if
// a service wants a special logger.

var Log = log.WithFields(DefaultFields)

// Add the 'error' field w/ the error object to the Log Entry.
func LogErrField(err error) *log.Entry {
	return Log.WithField("error", err)
}

// return current level of standard logger
func GetLogLevel() log.Level {
	return log.GetLevel()
}

func InitializeOcelog(log_level string) {
	if loglevel, err := log.ParseLevel(log_level); err != nil {
		LogErrField(err).Fatal()
	} else {
		log.SetLevel(loglevel)
	}
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

func GetFlags() string {
    // write flag
    var log_level string
    flag.StringVar(&log_level, "log_level", "warn", "set log level")
    flag.Parse()
    return log_level
}

/*
---------
this is for the fancy function call stuff
*/
func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}

var LogrusPackage string
var RuntimePackage string

// Positions in the call stack when tracing to report the calling method
var minimumCallerDepth int

const maximumCallerDepth int = 25
const knownLogrusFrames int = 4

// I took and modified this from an ummerged PR for logrus
// getCaller retrieves the name of the first non-logrus / ocelog calling function
func getCaller() (method string) {
	// Restrict the lookback frames to avoid runaway lookups
	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)

	// cache this package's fully-qualified name
	if LogrusPackage == "" {
		LogrusPackage = getPackageName(runtime.FuncForPC(pcs[0]).Name())

		// now that we have the cache, we can skip a minimum count of known-logrus functions
		minimumCallerDepth = knownLogrusFrames
	}

	for i := 0; i < depth; i++ {
		fullFuncName := runtime.FuncForPC(pcs[i]).Name()
		pkg := getPackageName(fullFuncName)
		// If the caller isn't part of this package, we're done
		if pkg != LogrusPackage && !strings.Contains(fullFuncName, "ocelog") && !strings.Contains(fullFuncName, "getCaller") {
			return fullFuncName
		}
	}

	// if we got here, we failed to find the caller's context
	return ""
}
