/* Package ocelog
Way to have one style of logging for the project. Initialize in your service w/ InitializeOcelog(), uses a JSONFormatter.
use ocelog.Log() to log with extra field of the function called.

todo: add common log functions, right now there is only IncludeErrField which adds the error: <error text> to the json log.

*/
package ocelog

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strings"
)

// TODO: add NewLog() function that returns a logger w/ the default fields, if  a service wants a special logger.

// Log() is the default logger for Ocelog. Includes extra field `"function": <name>`
func Log() *log.Entry {
	return log.WithFields(GetDefaultFields())
}

// IncludeErrField adds the 'error' field w/ the error object to the Log Entry. This still requires setting a
// Info/warning/w/e message
// Example:
//   ocelog.IncludeErrField(err).Error("booo code made an error")
func IncludeErrField(err error) *log.Entry {
	return Log().WithField("error", err)
}

// Adds "function" field
func GetDefaultFields() log.Fields {
	return log.Fields{
		"function": getCaller(),
	}
}

// GetLogLevel returns the current level of standard logger.
func GetLogLevel() log.Level {
	return log.GetLevel()
}


// configure default Logger to log in JSON format.
func InitializeOcelog(logLevel string) {
	if loglevel, err := log.ParseLevel(logLevel); err != nil {
		IncludeErrField(err).Fatal()
	} else {
		log.SetLevel(loglevel)
	}
	log.SetFormatter(&log.JSONFormatter{
		//PrettyPrint: true,
	})
	log.SetOutput(os.Stdout)
}

// GetFlags() gets log level flags from command line.
// ex:
// `hookhandler --log_level=debug` */
func GetFlags() string {
	// write flag
	var logLevel string
	flag.StringVar(&logLevel, "log_level", "warn", "set log level")
	flag.Parse()
	return logLevel
}

// GetPackageName returns pkg name of calling function.
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

// for caching of Logrus package when searching for calling function
var LogrusPackage string

// Positions in the call stack when tracing to report the calling method
var minimumCallerDepth int

// limit to search depth
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
