package sendstream

import (
	"fmt"

	"github.com/level11consulting/orbitalci/build/streaminglogs"
	models "github.com/level11consulting/orbitalci/models/pb"
	"github.com/shankj3/go-til/log"
)

type Sendy interface {
	Send(response *models.LineResponse) error
}

//SendStream will send a message formatted by strFmt string with variables fmtVars... will log error if it finds it
func SendStream(sendy Sendy, strFmt string, fmtVars ...interface{}) {
	if err := sendy.Send(streaminglogs.RespWrap(fmt.Sprintf(strFmt, fmtVars...))); err != nil {
		log.IncludeErrField(err).Error("error sending stream")
	}
}
