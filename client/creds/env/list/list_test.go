package envlist

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/shankj3/ocelot/models/pb"
)

var creds = []*pb.GenericCreds{
	{
		AcctName: "one",
		Identifier: "identifier1",
		ClientSecret: "secret1",
		SubType: pb.SubCredType_ENV,
	},
	{
		AcctName: "one",
		Identifier: "identifier2",
		ClientSecret: "secret2",
		SubType: pb.SubCredType_ENV,
	},
	{
		AcctName: "two",
		Identifier: "identifier3",
		ClientSecret: "secret3",
		SubType: pb.SubCredType_ENV,
	},
	{
		AcctName: "one",
		Identifier: "identifier4",
		ClientSecret: "secret4",
		SubType: pb.SubCredType_ENV,
	},
}

func Test_organize(t *testing.T) {
	organized := organize(&pb.GenericWrap{Creds:creds})
	onecreds, ok := organized["one"]
	if !ok {
		t.Error("should have an array of env creds under acctname 'one'")
	}
	knownCreds := []*pb.GenericCreds{creds[0], creds[1], creds[3]}
	if diff := deep.Equal(knownCreds, *onecreds); diff != nil {
		t.Error(diff)
	}
	twocreds, ok := organized["two"]
	if !ok {
		t.Error("should have an array of env creds under acctname 'two'")
	}
	known2Creds := []*pb.GenericCreds{creds[2]}
	if diff := deep.Equal(known2Creds, *twocreds); diff != nil {
		t.Error(diff)
	}
}

func BenchmarkOrganizeCred10(b *testing.B) {
	var credz []*pb.GenericCreds
	for n := 0; n < 10; n++ {
		credz = append(credz, creds...)
	}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		organize(&pb.GenericWrap{Creds:credz})
	}
}


func BenchmarkOrganizeCred100(b *testing.B) {
	var credz []*pb.GenericCreds
	for n := 0; n < 100; n++ {
		credz = append(credz, creds...)
	}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		organize(&pb.GenericWrap{Creds:credz})
	}
}


func organize2(cred *pb.GenericWrap) map[string][]*pb.GenericCreds {
	organizedCreds := make(map[string][]*pb.GenericCreds)
	for _, cred := range cred.Creds {
		acctCreds := organizedCreds[cred.AcctName]
		acctCreds = append(acctCreds, cred)
		organizedCreds[cred.AcctName] = acctCreds
	}
	return organizedCreds
}

func BenchmarkOrganize2Cred10(b *testing.B) {
	var credz []*pb.GenericCreds
	for n := 0; n < 10; n++ {
		credz = append(credz, creds...)
	}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		organize2(&pb.GenericWrap{Creds:credz})
	}
}


func BenchmarkOrganize2Cred100(b *testing.B) {
	var credz []*pb.GenericCreds
	for n := 0; n < 100; n++ {
		credz = append(credz, creds...)
	}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		organize2(&pb.GenericWrap{Creds:credz})
	}
}
