package valet

func ConsulConnectionErr(err error) {
	msg := "Consul connection error! Cannot continue! Error: "
	panic(msg + err.Error())
}