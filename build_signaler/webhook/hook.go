package webhook


//	"io"
//
//	"github.com/shankj3/go-til/deserialize"
//	"github.com/shankj3/ocelot/common/credentials"
//	"github.com/shankj3/ocelot/common/remote"
//	"github.com/shankj3/ocelot/models"
//	"github.com/shankj3/ocelot/storage"
//)

// stub? i guess?

// HERE PUT FUNCTION FOR CREATING COMMIT LIST FROM PUSH

// HERE PUT FUNCTION FOR CREATING COMMIT LIST FROM PULL REQUEST


//cfg, err := credentials.GetVcsCreds(w.Store, w.AcctRepo, w.RC)
//if err != nil {
//return errors.New("couldn't get vcs creds, error: " + err.Error())
//}
//handler, token, err := remote.GetHandler(cfg)
//if err != nil {
//return errors.New("could not get remote client, error: " + err.Error())
//}
//w.handler = handler
////w.token = token
////return nil
//
//func GetAuth(store storage.CredTable, acctRepo string, cred credentials.CVRemoteConfig) (models.VCSHandler, string, error) {
//	cfg, err := credentials.GetVcsCreds(store, acctRepo, cred)
//	if err != nil {
//		return nil, "", err
//	}
//	return remote.GetHandler(cfg)
//}
//
//func PushRecieve(reciever models.HookReceiver, dese *deserialize.Deserializer, body io.Reader) error {
//	push, err := reciever.TranslatePush(body)
//	if err != nil {
//		return err
//	}
//	push.Repo.AcctRepo
//
//}