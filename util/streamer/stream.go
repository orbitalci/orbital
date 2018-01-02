package streamer

import (
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bufio"
	"bytes"
	"fmt"
	"time"
)

// StreamFromArray will do exactly that, it will write to the Streamable stream from an array.
// it will wait 0.1s before checking array again for new content
func StreamFromArray(array StreamArray, stream Streamable, debug func(...interface{})) (err error) {
	var index int
	var previousIndex int
	for {
		time.Sleep(100) // todo: set polling to be configurable
		fullArrayStreamed := len(array.GetData()) == index
		if array.CheckDone() && fullArrayStreamed {
			debug("done streaming from array")
			return nil
		}
		// if no new data has been sent, don't even try
		if fullArrayStreamed {
			continue
		}
		dataArray := array.GetData()[index:]
		ind, err := iterateOverByteArray(dataArray, stream)
		previousIndex = index
		index += ind
		debug(fmt.Sprintf("lines sent: %s | index: %s | previousIndex: %s", ind, index, previousIndex))
		if err != nil {
			return err
		}
	}

}


func iterateOverByteArray(data [][]byte, stream Streamable) (int, error) {
	var index int
	for ind, dataLine := range data {
		if err := stream.SendIt(dataLine); err != nil {
			return ind, err
		}
		// adding the number of lines added to index so streamFromArray knows where to start on the next pass
		index = ind + 1
	}
	return index, nil
}

func StreamFromStorage(store storage.BuildOutputStorage, stream Streamable, storageKey string) error {
	bytez, err := store.Retrieve(storageKey)
	if err != nil {
		stream.SendError([]byte("could not retrieve persisted build data"))
	}
	reader := bytes.NewReader(bytez)
	s := bufio.NewScanner(reader)
	for s.Scan() {
		if err := stream.SendIt(s.Bytes()); err != nil {
			return err
		}
	}
	return s.Err()
}
