package webhook

import (
	"io"

	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/ocelot/models"
)

// stub? i guess?

// HERE PUT FUNCTION FOR CREATING COMMIT LIST FROM PUSH

// HERE PUT FUNCTION FOR CREATING COMMIT LIST FROM PULL REQUEST


func PushRecieve(reciever models.HookReceiver, dese *deserialize.Deserializer, body io.Reader) error {
	push, err := reciever.TranslatePush(body)
	if err != nil {
		return err
	}

}