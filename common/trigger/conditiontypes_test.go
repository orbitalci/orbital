package trigger

import (
	"testing"

	"github.com/go-test/deep"
)

func TestBranchCondition(t *testing.T) {
	brch := &BranchCondition{}
	if brch.GetTriggerType() != Branch {
		t.Error("branch condition needs trigger type of branch")
	}
	brch.SetLogical(Or)
	if brch.GetLogical() != Or {
		t.Error("logical should be or, since it was set to or")
	}
	brch.AddConditionValue("here")
	brch.AddConditionValue("there")
	if diff := deep.Equal(brch.GetConditionValues(), []string{"here", "there"}); diff != nil {
		t.Error(diff)
	}
	triggerData := &TriggerData{branch: "borg"}
	if brch.PassesMuster(triggerData) {
		t.Error("shouldn't pass, as 'borg' is not in the list")
	}
	triggerData.branch = "there"
	if !brch.PassesMuster(triggerData) {
		t.Error("should pass, as 'there' is in the condition list")
	}
}

func Test_stringInChangeList(t *testing.T) {
	changelist := []string{"hummuna bummuna", "[sup_dog] testing here but should not trigger"}
	if stringInChangeList("sup_dogs", changelist) {
		t.Error("only sup_dog is in the changelist, not sup_dogs. should not return true")
	}
	changelist = append(changelist, "[sup_dogs] trigger here")
	if !stringInChangeList("sup_dogs", changelist) {
		t.Error("sup_dogs is now in the changelist, should return true")
	}
}

func TestTextCondition(t *testing.T) {
	text := &TextCondition{}
	if text.GetTriggerType() != Text {
		t.Error("trigger type shoudl be Text, got " + text.GetTriggerType().String())
	}
	text.SetLogical(And)
	if text.GetLogical() != And {
		t.Error("set logical to And, should return And")
	}
	conditionals := []string{"text1", "text2345", "t82nsla7812l"}
	for _, condition := range conditionals {
		text.AddConditionValue(condition)
	}
	if diff := deep.Equal(conditionals, text.GetConditionValues()); diff != nil {
		t.Error(diff)
	}
	td := &TriggerData{commitTexts: []string{"text1 goes here and that's real nice", "text2345 is also here!!!", "but we don't have the last nonsensical boi"}}
	if text.PassesMuster(td) {
		t.Error("should not pass muster, it doens't have t82nsla7812l")
	}

	text.SetLogical(Or)
	if !text.PassesMuster(td) {
		t.Error("logical is Or and 2/3 messages are here, this should pass")
	}
	text.SetLogical(And)

	td.commitTexts = append(td.commitTexts, "now we got it.... t82nsla7812l")
	if !text.PassesMuster(td) {
		t.Error("should  pass muster, it now has t82nsla7812l")
	}

	if !text.PassesMuster(&TriggerData{commitTexts: []string{"text1 geez text2345 t82nsla7812l"}}) {
		t.Error("should pass, as it has all the commit text messages")
	}
}

func TestFilepathCondition(t *testing.T) {
	fp := &FilepathCondition{}
	if fp.GetTriggerType() != Filepath {
		t.Error("trigger type should be filepath")
	}
	fp.SetLogical(Or)
	if fp.GetLogical() != Or {
		t.Error("logical was set to or, should be Or")
	}
	conditionals := []string{"src/test", "src/test/resources", "src/main/java", "src/main/resources"}
	for _, condition := range conditionals {
		fp.AddConditionValue(condition)
	}

	if diff := deep.Equal(conditionals, fp.GetConditionValues()); diff != nil {
		t.Error(diff)
	}
	changelist := []string{"src/test/garbo", "src/test/resource/somexml.xml", "src/main/jarra"}
	if !fp.PassesMuster(&TriggerData{filesChanged: changelist}) {
		t.Error("should pass because logical is Or and src/test is in the changelist")
	}
	fp.logical = And
	if fp.PassesMuster(&TriggerData{filesChanged: changelist}) {
		t.Error("logical is And, not all conditions are present. this should fail")
	}

	robustList := []string{"src/test/resources/main.yaml", "src/main/java/com/boop/here.java", "src/main/resources/com/boop/here.xml"}
	if !fp.PassesMuster(&TriggerData{filesChanged: robustList}) {
		t.Error("all required changes are present, this should pass")
	}
}
