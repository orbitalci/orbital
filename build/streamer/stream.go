package streamer

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shankj3/ocelot/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"time"
)

var (
	streamErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ocelot_werker_stream_errors_total",
			Help: "streaming errors from werker streaming build logs",
		},
	)
)

func init() {
	prometheus.MustRegister(streamErrors)
}

type StreamCancelled struct {
	message string
}

func (sc *StreamCancelled) Error() string {
	return sc.message
}

func NewStreamCancelled(msg string) *StreamCancelled {
	return &StreamCancelled{message: msg}
}

// StreamFromArray will do exactly that, it will write to the Streamable stream from an array.
// it will wait 0.1s before checking array again for new content
func StreamFromArray(array StreamArray, stream Streamable, log Loggy) (err error) {
	var index int
	//var previousIndex int
	for {
		time.Sleep(100) // todo: set polling to be configurable
		numLines := len(array.GetData())
		fullArrayStreamed := numLines == index
		if array.CheckDone() && fullArrayStreamed {
			log.Info("done streaming from array")
			return
		}
		// if no new data has been sent, don't even try
		if fullArrayStreamed {
			//debug("full array streamed")
			continue
		}
		if index > numLines {
			streamErrors.Inc()
			log.Error(fmt.Sprintf("WOAH THERE. SOMETHING IS STUPID WRONG. Index: %d, numLines: %d", index, numLines))
			continue
		}
		//dataArray := array.GetData()[index:]
		//previousIndex = index
		index, err = iterateOverByteArray(array, stream, index)
		if err != nil {
			if _, ok := err.(*StreamCancelled); !ok {
				streamErrors.Inc()
				log.Error("Error! ", err.Error(), " arraylength: ", strconv.Itoa(len(array.GetData())))
			} else {
				// if the stream was cancelled by the client, that's alright. we don't care about that.
				log.Info("ocelot client cancelled stream")
				err = nil
			}
			return
		}
	}
	return
}

func iterateOverByteArray(array StreamArray, stream Streamable, index int) (int, error) {
	array.Lock()
	defer array.Unlock()
	data := array.GetData()[index:]
	for _, dataLine := range data {
		if dataLine == nil {
			fmt.Println("WHY!!!!!!!")
			return index, errors.New("data line was nil! how tf")
		}
		if err := stream.SendIt(dataLine); err != nil {
			stat, _ := status.FromError(err)
			// if the operation was cancelled by the client, don't report an error
			if stat.Code() == codes.Canceled {
				return index, NewStreamCancelled("cancelled by client")
			}
			return index, err
		}
		// adding the number of lines added to index so streamFromArray knows where to start on the next pass
		index += 1
	}
	return index, nil
}

//StreamFromStorage will retrieve the entire build output from storage, send over stream line by line using *bufio.Scanner
func StreamFromStorage(store storage.BuildOut, stream Streamable, storageKey int64) error {
	output, err := store.RetrieveOut(storageKey)
	if err != nil {
		stream.SendError([]byte("could not retrieve persisted build data"))
	}
	reader := bytes.NewReader(output.Output)
	s := bufio.NewScanner(reader)
	for s.Scan() {
		if err := stream.SendIt(s.Bytes()); err != nil {
			return err
		}
	}
	return s.Err()
}
