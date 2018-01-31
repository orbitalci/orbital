package streamer

import (
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// StreamFromArray will do exactly that, it will write to the Streamable stream from an array.
// it will wait 0.1s before checking array again for new content
func StreamFromArray(array StreamArray, stream Streamable, debug func(...interface{})) (err error) {
	var index int
	//var previousIndex int
	for {
		time.Sleep(100) // todo: set polling to be configurable
		numLines := len(array.GetData())
		fullArrayStreamed := numLines == index
		if array.CheckDone() && fullArrayStreamed {
			debug("done streaming from array")
			return
		}
		// if no new data has been sent, don't even try
		if fullArrayStreamed {
			//debug("full array streamed")
			continue
		}
		if index > numLines {
			debug(fmt.Sprintf("WOAH THERE. SOMETHING IS STUPID WRONG. Index: %d, numLines: %d", index, numLines))
			continue
		}
		//dataArray := array.GetData()[index:]
		//previousIndex = index
		index, err = iterateOverByteArray(array, stream, index)
		if err != nil {
			debug("ERROR! " + err.Error())
			debug("len array is: ", strconv.Itoa(len(array.GetData())))
			return
			//return err
		}
		//debug(fmt.Sprintf("lines sent: %d | index: %d | previousIndex: %d | length: %d", index - previousIndex, index, previousIndex, len(array.GetData())))
	}
	return
}


func iterateOverByteArray(array StreamArray, stream Streamable, index int) (int, error) {
	array.Lock()
	data := array.GetData()[index:]
	for _, dataLine := range data {
		if dataLine == nil {
			fmt.Println("WHY!!!!!!!")
			return index, errors.New("data line was nil! how tf")
		}
		if err := stream.SendIt(dataLine); err != nil {
			return index, err
		}
		// adding the number of lines added to index so streamFromArray knows where to start on the next pass
		index += 1
	}
	array.Unlock()
	return index, nil
}

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
