package ocenet

import (
	"github.com/meatballhat/negroni-logrus"
	"github.com/shankj3/ocelot/util/ocelog"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"net/http"
)


// InitNegroni is a helper function for starting up http servers. it will create a new Negroni instance,
// use attach a logrus instance with a json formatter, and register the log instance with the appName
// it will also attach a handler, so really all you have to call is Run() on the returned instance
func InitNegroni(appName string, handler http.Handler) (n *negroni.Negroni){
	n = negroni.New(negroni.NewRecovery(), negroni.NewStatic(http.Dir("public")))
	n.Use(negronilogrus.NewCustomMiddleware(ocelog.GetLogLevel(), &logrus.JSONFormatter{}, appName))
	n.UseHandler(handler)
	return
}

// for initializing connections / configurations once, then passing it around for lifetime of application
// Ctx can be anything, just have to set H to be HandleFunc that also takes in the context as first value.
// in handle func, cast Ctx interface{} to your struct that you initialized in startup so you can access fields
// for ex:
// ```
// appctx := &MyContext{config: "config yay wooo the best"}
// muxi.Handle("/ws/builds/{hash}", &ocenet.AppContextHandler{appctx, stream}).Methods("GET")
// ...
//
// func stream(ctx interface{}, w http.ResponseWriter, r *http.Request){
//     a := ctx.(*MyContext)
//     // do stuff
// ...
// ```
type AppContextHandler struct {
	Ctx interface{}
	H func(interface{}, http.ResponseWriter, *http.Request)
}


func (ah *AppContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ah.H(ah.Ctx, w, r)
}
