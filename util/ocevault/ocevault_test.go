package ocevault


//
//import (
//"fmt"
//"testing"
//)
////
//// THESE TESTS ARE DEPENDENT ON VAULT INSTANCE BEING UP, VAULT_ADDR BEING SET APPROPRIATELY,
//// AND VAULT_TOKEN BEING SET.
////
//func TestOcevault_CreateThrowawayToken(t *testing.T) {
//	oce, err := NewEnvAuthClient()
//	if err != nil {
//		t.Fatal(err)
//	}
//	secret, err := oce.CreateThrowawayToken()
//	if err != nil {
//		t.Fatal(err)
//	}
//	fmt.Println("Token:", secret)
//}
//
//func TestOcevault_GetUserAuthData(t *testing.T) {
//	oce, err := NewEnvAuthClient()
//	if err != nil {
//		t.Fatal(err)
//	}
//	sec, err := oce.GetUserAuthData("jessi")
//	if err != nil {
//		t.Fatal(err)
//	}
//	fmt.Println("ayyyy: ", sec)
//
//}
//
//func TestOcevault_CreateOcevaultPolicy(t *testing.T) {
//	oce, err := NewEnvAuthClient()
//	if err != nil {
//		t.Fatal(err)
//	}
//	if err = oce.CreateOcevaultPolicy(); err != nil {
//		t.Fatal(err)
//	}
//}