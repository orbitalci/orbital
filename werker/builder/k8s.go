package builder

type K8 struct {
}

func NewK8Builder() *K8 {
	return &K8{}
}

// just examples of what this can be, test

//func (d *K8Proc) RunPushBundle(bund *protos.PushBuildBundle, infochan chan []byte) {
//	ocelog.Log().Debug("building building tasty tasty push bundle")
//	// run push bundle.
//	//fmt.Println(bund.PushData.Repository.FullName)
//	infochan <- []byte(bund.PushData.Repository.FullName)
//	infochan <- []byte(bund.PushData.Repository.Owner.Username)
//	infochan <- []byte("gonna sleep for 5 seconds now.")
//	time.Sleep(5 * time.Second)
//	infochan <- []byte("push requeeeeeeeest KUBERNETES!")
//	infochan <- []byte("sleeping for 5 more seconds!!!!")
//	time.Sleep(5 * time.Second)
//	infochan <- []byte("this could be some delightful std out from builds! huzzah! I'M RUNNING W/ KUBERNETES!")
//	close(infochan)
//}
//
//func (d *K8Proc) RunPRBundle(bund *protos.PRBuildBundle, infochan chan []byte) {
//	infochan <- []byte(bund.PrData.Repository.FullName)
//	infochan <- []byte("delightful! KUBERNETES! love KUBERNETES!")
//	infochan <- []byte("kubernetes pulllll reqquuuueeeeeeeeeest!")
//	close(infochan)
//}
