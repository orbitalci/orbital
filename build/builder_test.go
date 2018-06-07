package build

import (
	"testing"

	"github.com/shankj3/go-til/test"
)

func TestInitStageUtil(t *testing.T) {
	stage := InitStageUtil("sTg")
	if stage.Stage != "stg" {
		t.Error(test.StrFormatErrors("stage", "stg", stage.Stage))
	}
}

func TestStageUtil(t *testing.T) {
	su := InitStageUtil("sttTg")
	if su.GetStage() != "stttg" {
		t.Error(test.StrFormatErrors("stage", "sttg", su.GetStage()))
	}
	if su.GetStageLabel() != "STTTG | " {
		t.Error(test.StrFormatErrors("stage label", "STTTG | ", su.GetStageLabel()))
	}
	su.SetStage("scrrrtg")
	if su.GetStage() != "scrrrtg" {
		t.Error(test.StrFormatErrors("setted stage", "scrrrtg", su.GetStage()))
	}
	su.SetStageLabel("COWABUNUGA")
	if su.GetStageLabel() != "COWABUNUGA" {
		t.Error(test.StrFormatErrors("setted stage label", "COWABUNUGA", su.GetStageLabel()))
	}
}

func TestCreateSubstage(t *testing.T) {
	su := InitStageUtil("subbed")
	subsu := CreateSubstage(su, "subber")
	if subsu.GetStageLabel() != "SUBBED | SUBBER | " {
		t.Error(test.StrFormatErrors("substage label", "SUBBED | SUBBER | ", subsu.GetStageLabel()))
	}
	if subsu.GetStage() != "subbed | subber" {
		t.Error(test.StrFormatErrors("substage name", "subbed | subber", subsu.GetStage()))
	}
}