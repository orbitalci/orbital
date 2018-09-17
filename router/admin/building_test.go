package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hashicorp/consul/api"
	"github.com/shankj3/go-til/consul"
	"github.com/shankj3/ocelot/common"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGuideOcelotServer_BuildRuntime_hash(t *testing.T) {
	consl := &buildruntimeconsl{}
	rc := &credentials.RemoteConfig{Consul: consl}
	gos := &guideOcelotServer{RemoteConfig: rc, Storage: &buildruntimestorage{}}
	bilds, err := gos.BuildRuntime(context.Background(), &pb.BuildQuery{Hash: "1234sdfasdfasf"})
	if err != nil {
		t.Error(err)
		return
	}
	if bilds.Builds["1234sdfasdfasf"].Done {
		t.Error("still in consul, this buildruntime should be Done:false")
	}
	consl = &buildruntimeconsl{empty: true}
	rc = &credentials.RemoteConfig{Consul: consl}
	gos = &guideOcelotServer{RemoteConfig: rc, Storage: &buildruntimestorage{}}
	bilds, err = gos.BuildRuntime(context.Background(), &pb.BuildQuery{Hash: "1234sdfasdfasf"})
	if err != nil {
		t.Error(err)
		return
	}
	if !bilds.Builds["1234sdfasdfasf"].Done {
		t.Error("build not in consul, this buildru/ntime shoudl be Done:true")
	}
	consl.fail = true
	bilds, err = gos.BuildRuntime(context.Background(), &pb.BuildQuery{Hash: "1234sdfasdfasf"})
	if bilds != nil || err == nil {
		t.Error("consul 'failed', this shoud return an error and a nil summary obj")
	}
	consl = &buildruntimeconsl{}
	rc.Consul = consl
	gos = &guideOcelotServer{RemoteConfig: rc, Storage: &buildruntimestorage{fail: true}}
	bilds, err = gos.BuildRuntime(context.Background(), &pb.BuildQuery{Hash: "1234sdfasdfasf"})
	if err == nil {
		t.Error("storage returned an error, this should bubble up that error")
	}
	if grpcErr, ok := status.FromError(err); !ok {
		t.Error("should return a grpc error")
	} else if grpcErr.Code() != codes.Internal {
		t.Error("storage did not return a not found, this code should be internal")
	}
	consl = &buildruntimeconsl{}
	rc.Consul = consl
	gos = &guideOcelotServer{RemoteConfig: rc, Storage: &buildruntimestorage{notFound: true}}
	bilds, err = gos.BuildRuntime(context.Background(), &pb.BuildQuery{Hash: "1234sdfasdfasf"})
	if err == nil {
		t.Error("storage returned an error, this should bubble up that error")
	}
	if grpcErr, ok := status.FromError(err); !ok {
		t.Error("should return a grpc error")
	} else if grpcErr.Code() != codes.NotFound {
		t.Error("storage returned a not found, grpc code should also be not found. instead its " + grpcErr.Code().String())
	}
}
func TestGuideOcelotServer_BuildRuntime_none(t *testing.T) {
	g := &guideOcelotServer{}
	_, err := g.BuildRuntime(context.Background(), &pb.BuildQuery{})
	if err == nil {
		t.Error("no arguments passed, shoudl return validation error")
	}
}

func TestGuideOcelotServer_BuildRuntime_buildId(t *testing.T) {
	consl := &buildruntimeconsl{}
	rc := &credentials.RemoteConfig{Consul: consl}
	gos := &guideOcelotServer{RemoteConfig: rc, Storage: &buildruntimestorage{}}
	bilds, err := gos.BuildRuntime(context.Background(), &pb.BuildQuery{BuildId: 12})
	if err != nil {
		t.Error(err)
	}
	if !bilds.Builds["1234"].Done {
		t.Error("builds retrieved by id return should return done:true")
	}
	if bilds.Builds["1234"].AcctName != "shankj3" {
		t.Error("wtf")
	}
	gos = &guideOcelotServer{RemoteConfig: rc, Storage: &buildruntimestorage{fail: true}}
	bilds, err = gos.BuildRuntime(context.Background(), &pb.BuildQuery{BuildId: 12})
	if err == nil {
		t.Error("storage failed, should return error")
	}
}

func TestGuideOcelotServer_Logs(t *testing.T) {
	g := &guideOcelotServer{}
	err := g.Logs(&pb.BuildQuery{}, nil)
	if err == nil {
		t.Error("empty query, hsould return failure")
	}
	consl := &buildruntimeconsl{}
	rc := &credentials.RemoteConfig{Consul: consl}
	gos := &guideOcelotServer{RemoteConfig: rc, Storage: &buildruntimestorage{}}
	logserver := &logserv{}
	err = gos.Logs(&pb.BuildQuery{BuildId: 12}, logserver)
	if err != nil {
		t.Error(err)
	}
	outlines := strings.Split(output, "\n")
	if diff := deep.Equal(outlines[:10], logserver.sentLines); diff != nil {
		t.Error(diff)
	}
	gos = &guideOcelotServer{RemoteConfig: rc, Storage: &buildruntimestorage{fail: true}}
	if err := gos.Logs(&pb.BuildQuery{BuildId: 12}, nil); err == nil {
		t.Error("should have failed, storage returned a failure")
	} else if !strings.Contains(err.Error(), "Unable to retrive from FAKE") {
		t.Error("should bubble up dstorage failure, instead: " + err.Error())
	}
}

func TestGuideOcelotServer_Logs_fromHash(t *testing.T) {
	consl := &buildruntimeconsl{}
	rc := &credentials.RemoteConfig{Consul: consl}
	gos := &guideOcelotServer{RemoteConfig: rc, Storage: &buildruntimestorage{}}
	serv := &logserv{}
	err := gos.Logs(&pb.BuildQuery{Hash: "1234"}, serv)
	if err == nil {
		t.Error("build in consul, should return an error")
	}
	if serv.sentLines[0] != "build is not finished, use BuildRuntime method and stream from the werker registered" {
		t.Error("should return build not finished in stream ")
	}
	consl = &buildruntimeconsl{empty: true}
	serv = &logserv{}
	rc = &credentials.RemoteConfig{Consul: consl}
	gos = &guideOcelotServer{RemoteConfig: rc, Storage: &buildruntimestorage{}}
	err = gos.Logs(&pb.BuildQuery{Hash: "1234"}, serv)
	if err != nil {
		t.Error("should not fail, error is: " + err.Error())
	}
	expectedoutlines := strings.Split(output, "\n")
	if diff := deep.Equal(expectedoutlines[:10], serv.sentLines); diff != nil {
		t.Error(diff)
	}
}

func TestGuideOcelotServer_FindWerker(t *testing.T) {
	consl := &buildruntimeconsl{empty: true}
	rc := &credentials.RemoteConfig{Consul: consl}
	gos := &guideOcelotServer{RemoteConfig: rc, Storage: &buildruntimestorage{}}
	if _, err := gos.FindWerker(context.Background(), &pb.BuildReq{Hash: "1234"}); err == nil {
		t.Error("shouldreturn error, as consul is empty")
	} else if !strings.Contains(err.Error(), "werker not found for request as it has already finished") {
		t.Error("error should be build finished, got: " + err.Error())
	}
	consl.empty = false
	brtinfo, err := gos.FindWerker(context.Background(), &pb.BuildReq{Hash: "1234"})
	if err != nil {
		t.Error("shouldn't fail")
	}
	if brtinfo.Hash != "1234" {
		t.Error("returned wrong info")
	}
	consl.empty = false
	consl.fail = true
	_, err = gos.FindWerker(context.Background(), &pb.BuildReq{Hash: "1234"})
	if err == nil {
		t.Error("consul returned an unknown error, should bubble up")
	}
	if !strings.Contains(err.Error(), "could not get build runtime, err: ") {
		t.Error("should return build runtime not found error, instead got: " + err.Error())
	}
	consl.fail = false
	consl.runtimeMulti = true
	_, err = gos.FindWerker(context.Background(), &pb.BuildReq{Hash: "1234"})
	if err == nil {
		t.Error("multiple werker builds found for hash, should return an error")
		return
	}
	if err.Error() != "rpc error: code = InvalidArgument desc = ONE and ONE ONLY match should be found for your hash" {
		t.Error("should return multiple hash match error, got: " + err.Error())
	}
	_, err = gos.FindWerker(context.Background(), &pb.BuildReq{})
	if err == nil {
		t.Error("no hash passed, should return an error")
	}
	if err.Error() != "rpc error: code = InvalidArgument desc = Please pass a hash" {
		t.Error("should return hash validation error")
	}

}

type buildruntimeconsl struct {
	consul.Consuletty
	fail         bool
	empty        bool
	runtimeMulti bool
	uuid         string
	hash         string
}

func (b *buildruntimeconsl) GetKeyValue(path string) (*api.KVPair, error) {
	if b.fail {
		return nil, errors.New("failing at GetKeyValue")
	}
	if b.empty {
		return nil, nil
	}
	return &api.KVPair{Key: path, Value: []byte("uuiuuiuuiuuiduiduiduid")}, nil

}

func (b *buildruntimeconsl) GetKeyValues(prefix string) (api.KVPairs, error) {
	var kvp api.KVPairs
	if b.fail {
		return nil, errors.New("failin")
	}
	if b.empty {
		return api.KVPairs{}, nil
	}
	fullHash := common.ParseBuildMapPath(prefix)
	if strings.Contains(prefix, "werker_build_map") {
		if b.runtimeMulti {
			kvp = api.KVPairs{{Key: "ci/werker_build_map/" + fullHash, Value: []byte(b.uuid)},
				{Key: "ci/werker_build_map/" + fullHash + "2", Value: []byte(b.uuid + "2")}}
		} else {
			kvp = api.KVPairs{{Key: "ci/werker_build_map/" + fullHash, Value: []byte(b.uuid)}}
		}
		return kvp, nil
	}
	if strings.Contains(prefix, "ci/werker_location/") {
		kvp = api.KVPairs{
			{Key: "ci/werker_location/" + b.uuid + "/werker_ip", Value: []byte("localhost")},
			{Key: "ci/werker_location/" + b.uuid + "/werker_grpc_port", Value: []byte("9090")},
			{Key: "ci/werker_location/" + b.uuid + "/werker_ws_port", Value: []byte("9099")},
		}
		return kvp, nil
	}
	return nil, nil
}

type buildruntimestorage struct {
	storage.OcelotStorage
	notFound    bool
	fail        bool
	buildFailed bool
	hash        string
}

func (b *buildruntimestorage) RetrieveHashStartsWith(partialGitHash string) ([]*pb.BuildSummary, error) {
	if b.notFound {
		return nil, storage.BuildSumNotFound(partialGitHash)
	}
	if b.fail {
		return nil, errors.New("failing storage at RetrieveHashStartsWith")
	}
	return []*pb.BuildSummary{{Hash: partialGitHash, Failed: b.buildFailed, QueueTime: &timestamp.Timestamp{Seconds: time.Now().Add(-time.Hour).Unix()}, BuildTime: &timestamp.Timestamp{Seconds: time.Now().Add(-time.Hour).Unix()}, Account: "shankj3", Repo: "ocelot", Branch: "master", BuildId: 1}}, nil

}

func (b *buildruntimestorage) RetrieveLatestSum(gitHash string) (*pb.BuildSummary, error) {
	if b.notFound {
		return nil, storage.BuildSumNotFound(gitHash)
	}
	if b.fail {
		return nil, errors.New("failing storage at RetrieveLatestSum")
	}
	return &pb.BuildSummary{Hash: gitHash, Failed: b.buildFailed, QueueTime: &timestamp.Timestamp{Seconds: time.Now().Add(-time.Hour).Unix()}, BuildTime: &timestamp.Timestamp{Seconds: time.Now().Add(-time.Hour).Unix()}, Account: "shankj3", Repo: "ocelot", Branch: "master", BuildId: 12}, nil
}

func (b *buildruntimestorage) RetrieveSumByBuildId(buildId int64) (*pb.BuildSummary, error) {
	if b.notFound {
		return nil, storage.BuildSumNotFound(fmt.Sprintf("%d", buildId))
	}
	if b.fail {
		return nil, errors.New("failing storage at RetrieveSumByBuildId")
	}
	return &pb.BuildSummary{Hash: "1234", Failed: b.buildFailed, QueueTime: &timestamp.Timestamp{Seconds: time.Now().Add(-time.Hour).Unix()}, BuildTime: &timestamp.Timestamp{Seconds: time.Now().Add(-time.Hour).Unix()}, Account: "shankj3", Repo: "ocelot", Branch: "master", BuildId: buildId}, nil
}

var output = `ON an exceptionally hot evening early in July a young man came out of the garret in which he lodged in S. Place and walked slowly, as though in hesitation, towards K. bridge.	   1
  He had successfully avoided meeting his landlady on the staircase. His garret was under the roof of a high, five-storied house, and was more like a cupboard than a room. The landlady, who provided him with garret, dinners, and attendance, lived on the floor below, and every time he went out he was obliged to pass her kitchen, the door of which invariably stood open. And each time he passed, the young man had a sick, frightened feeling, which made him scowl and feel ashamed. He was hopelessly in debt to his landlady, and was afraid of meeting her.	   2
  This was not because he was cowardly and abject, quite the contrary; but for some time past he had been in an overstrained, irritable condition, verging on hypochondria. He had become so completely absorbed in himself, and isolated from his fellows that he dreaded meeting, not only his landlady, but any one at all. He was crushed by poverty, but the anxieties of his position had of late ceased to weigh upon him. He had given up attending to matters of practical importance; he had lost all desire to do so. Nothing that any landlady could do had a real terror for him. But to be stopped on the stairs, to be forced to listen to her trivial, irrelevant gossip, to pestering demands for payment, threats and complaints, and to rack his brains for excuses, to prevaricate, to lie—no, rather than that, he would creep down the stairs like a cat and slip out unseen.	   3
  This evening, however, on coming out into the street, he became acutely aware of his fears.	   4
  “I want to attempt a thing like that and am frightened by these trifles,” he thought, with an odd smile. “Hm … yes, all is in a man’s hands and he lets it all slip from cowardice, that’s an axiom. It would be interesting to know what it is men are most afraid of. Taking a new step, uttering a new word is what they fear most.… But I am talking too much. It’s because I chatter that I do nothing. Or perhaps it is that I chatter that I do nothing. I’ve learned to chatter this last month, lying for days together in my den thinking … of Jack the Giant-killer. Why am I going there now? Am I capable of that? Is that serious? It is not serious at all. It’s simply a fantasy to amuse myself; a plaything! Yes, maybe it is a plaything.”	   5
  The heat in the street was terrible: and the airlessness, the bustle and the plaster, scaffolding, bricks, and dust all about him, and that special Petersburg stench, so familiar to all who are unable to get out of town in summer—all worked painfully upon the young man’s already overwrought nerves. The insufferable stench from the pot-houses, which are particularly numerous in that part of the town, and the drunken men whom he met continually, although it was a working day, completed the revolting misery of the picture. An expression of the profoundest disgust gleamed for a moment in the young man’s refined face. He was, by the way, exceptionally handsome, above the average in height, slim, well-built, with beautiful dark eyes and dark brown hair. Soon he sank into deep thought, or more accurately speaking into a complete blankness of mind; he walked along not observing what was about him and not caring to observe it. From time to time, he would mutter something, from the habit of talking to himself, to which he had just confessed. At these moments he would become conscious that his ideas were sometimes in a tangle and that he was very weak; for two days he had scarcely tasted food.	   6
  He was so badly dressed that even a man accustomed to shabbiness would have been ashamed to be seen in the street in such rags. In that quarter of the town, however, scarcely any short-coming in dress would have created surprise. Owing to the proximity of the Hay Market, the number of establishments of bad character, the preponderance of the trading and working class population crowded in these streets and alleys in the heart of Petersburg, types so various were to be seen in the streets that no figure, however queer, would have caused surprise. But there was such accumulated bitterness and contempt in the young man’s heart that, in spite of all the fastidiousness of youth, he minded his rags least of all in the street. It was a different matter when he met with acquaintances or with former fellow students, whom, indeed, he disliked meeting at any time. And yet when a drunken man who, for some unknown reason, was being taken some-where in a huge wagon dragged by a heavy dray horse, suddenly shouted at him as he drove past: “Hey there, German hatter!” bawling at the top of his voice and pointing at him—the young man stopped suddenly and clutched tremulously at his hat. It was a tall round hat from Zimmerman’s, but completely worn out, rusty with age, all torn and bespattered, brimless and bent on one side in a most unseemly fashion. Not shame, however, but quite another feeling akin to terror had overtaken him.	   7
  “I knew it,” he muttered in confusion, “I thought so! That’s the worst of all! Why, a stupid thing like this, the most trivial detail might spoil the whole plan. Yes, my hat is too noticeable.… It looks absurd and that makes it noticeable.… With my rags I ought to wear a cap, any sort of old pancake, but not this grotesque thing. Nobody wears such a hat, it would be noticed a mile off, it would be remembered.… What matters is that people would remember it, and that would give them a clue. For this business one should be as little conspicuous as possible.… Trifles, trifles are what matter! Why, it’s just such trifles that always ruin everything.…”	   8
  He had not far to go; he knew indeed how many steps it was from the gate of his lodging house: exactly seven hundred and thirty. He had counted them once when he had been lost in dreams. At the time he had put no faith in those dreams and was only tantalising himself by their hideous but daring recklessness. Now, a month later, he had begun to look upon them differently, and, in spite of the monologues in which he jeered at his own impotence and indecision, he had involuntarily come to regard this “hideous” dream as an exploit to be attempted, although he still did not realise this himself. He was positively going now for a “rehearsal” of his project, and at every step his excitement grew more and more violent.	   9
  With a sinking heart and a nervous tremor, he went up to a huge house which on one side looked on to the canal, and on the other into the street. This house was let out in tiny tenements and was inhabited by working people of all kinds—tailors, locksmiths, cooks, Germans of sorts, girls picking up a living as best they could, petty clerks, &c. There 
`

func (b *buildruntimestorage) StorageType() string {
	return "FAKE"
}

func (b *buildruntimestorage) RetrieveOut(buildId int64) (models.BuildOutput, error) {
	if b.fail {
		return models.BuildOutput{}, errors.New("failing at retrieveOut")
	}
	return models.BuildOutput{BuildId: buildId, Output: []byte(output), OutputId: 1}, nil
}

func (b *buildruntimestorage) RetrieveLastOutByHash(gitHash string) (models.BuildOutput, error) {
	if b.fail {
		return models.BuildOutput{}, errors.New("failing at RetrieveLastOutByHash")
	}
	return models.BuildOutput{BuildId: 12, Output: []byte(output), OutputId: 1}, nil
}

type logserv struct {
	pb.GuideOcelot_LogsServer
	sentLines  []string
	returnErro bool
}

func (l *logserv) Send(resp *pb.LineResponse) error {
	if l.returnErro {
		return errors.New("failing at log send!")
	}
	l.sentLines = append(l.sentLines, resp.OutputLine)
	return nil
}
