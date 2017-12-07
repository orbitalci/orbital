package processors

import (
	"github.com/shankj3/ocelot/protos/out"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"time"
)

type DockProc struct {

}

func (d *DockProc) RunPushBundle(bund *protos.PushBuildBundle, infochan chan []byte) {
	ocelog.Log().Debug("building building tasty tasty push bundle")
	// run push bundle.
	//fmt.Println(bund.PushData.Repository.FullName)
	infochan <- []byte(bund.PushData.Repository.FullName)
	infochan <- []byte(bund.PushData.Repository.Owner.Username)
	infochan <- []byte("gonna sleep for 5 seconds now.")
	time.Sleep(5*time.Second)
	infochan <- []byte("push requeeeeeeeest DOCKER!")
	infochan <- []byte("sleeping for 5 more seconds!!!!")
	time.Sleep(4*time.Second)
	infochan <- []byte("this could be some delightful std out from builds! huzzah! I'M RUNNING W/ DOCKER!")
	close(infochan)
}

func (d *DockProc) RunPRBundle(bund *protos.PRBuildBundle, infochan chan []byte) {
	infochan <- []byte(bund.PrData.Repository.FullName)
	infochan <- []byte("delightful! docker! love docker!")
	infochan <- []byte("dockeeerrr pulllll reqquuuueeeeeeeeeest!")
	close(infochan)
}