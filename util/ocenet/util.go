package ocenet

import "net/http"

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

