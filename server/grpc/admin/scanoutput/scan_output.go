package scanoutput

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/shankj3/go-til/log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"regexp"
	"github.com/level11consulting/ocelot/server/grpc/admin/sendstream"
)

// Moved this regex in here because we only use it in this file
const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
const gradle = `.*<[=|-]{13}> [0-9]+% (CONFIGURING|EXECUTING|INITIALIZING|WAITING) \[([0-9]+m )?[0-9]+s\]>.*\n`

var re = regexp.MustCompile(ansi)
var regrad = regexp.MustCompile(gradle)

func maybeStrip(output []byte, stripAnsi bool) []byte {
	if stripAnsi {
		return regrad.ReplaceAll(re.ReplaceAll(output, []byte("")), []byte(""))
	}
	return output
}

// End regex copy/paste

// scanLog will create a scanner out of the buildOutput byte data and send it over the GuideOcelot logs stream.
//   will return a grpc error if something goes wrong
func ScanLog(out models.BuildOutput, stream pb.GuideOcelot_LogsServer, storageType string, stripAnsi bool) error {
	var cleanedOutput []byte
	if stripAnsi {
		cleanedOutput = maybeStrip(out.Output, stripAnsi)
	} else {
		cleanedOutput = out.Output
	}
	scanner := bufio.NewScanner(bytes.NewReader(cleanedOutput))
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		sendstream.SendStream(stream, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.IncludeErrField(err).Error("error encountered scanning from " + storageType)
		return status.Error(codes.Internal, fmt.Sprintf("Error was encountered while sending data from %s. \nError: %s", storageType, err.Error()))
	}
	return nil
}
