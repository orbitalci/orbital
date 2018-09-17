package trigger

import "strings"

type Section interface {
	GetTriggerType() TriggerType
	PassesMuster(*TriggerData) bool
	GetLogical() Conditional
	SetLogical(Conditional)
	AddConditionValue(string)
	GetConditionValues() []string
}

type BranchCondition struct {
	acceptedBranches []string
	logical          Conditional
}

func (b *BranchCondition) GetTriggerType() TriggerType {
	return Branch
}

func (b *BranchCondition) PassesMuster(td *TriggerData) bool {
	for _, branch := range b.acceptedBranches {
		if branch == td.branch {
			return true
		}
	}
	return false
}

func (b *BranchCondition) GetLogical() Conditional {
	return b.logical
}

func (b *BranchCondition) SetLogical(conditional Conditional) {
	b.logical = conditional
}

func (b *BranchCondition) AddConditionValue(str string) {
	b.acceptedBranches = append(b.acceptedBranches, str)
}

func (b *BranchCondition) GetConditionValues() []string {
	return b.acceptedBranches
}

func stringInChangeList(str string, stringz []string) bool {
	for _, strr := range stringz {
		if strings.Contains(strr, str) {
			return true
		}
	}
	return false
}

// changesPassMuster checks that all the values in a given changeset are acceptable under the rules defined by the accepted changes and the conditions (AND/OR) that combine them.
func changesPassMuster(logical Conditional, realChanges []string, acceptedChanges []string) bool {
	for _, change := range acceptedChanges {
		if logical == And {
			if !stringInChangeList(change, realChanges) {
				return false
			}
		} else if stringInChangeList(change, realChanges) {
			return true
		}
	}
	// if it is an AND condition and the loop has been exhausted w/ a match found every time,
	// then its good to go
	if logical == And {
		return true
	}
	return false
}

type TextCondition struct {
	acceptedTexts []string
	logical       Conditional
}

func (b *TextCondition) GetTriggerType() TriggerType {
	return Text
}

func (b *TextCondition) PassesMuster(td *TriggerData) bool {
	return changesPassMuster(b.logical, td.commitTexts, b.acceptedTexts)
}

func (b *TextCondition) GetLogical() Conditional {
	return b.logical
}

func (b *TextCondition) SetLogical(conditional Conditional) {
	b.logical = conditional
}

func (b *TextCondition) AddConditionValue(str string) {
	b.acceptedTexts = append(b.acceptedTexts, str)
}

func (b *TextCondition) GetConditionValues() []string {
	return b.acceptedTexts
}

///

type FilepathCondition struct {
	acceptedFilepaths []string
	logical           Conditional
}

func (b *FilepathCondition) GetTriggerType() TriggerType {
	return Filepath
}

func (b *FilepathCondition) PassesMuster(td *TriggerData) bool {
	return changesPassMuster(b.logical, td.filesChanged, b.acceptedFilepaths)
}

func (b *FilepathCondition) GetLogical() Conditional {
	return b.logical
}

func (b *FilepathCondition) SetLogical(conditional Conditional) {
	b.logical = conditional
}

func (b *FilepathCondition) AddConditionValue(str string) {
	b.acceptedFilepaths = append(b.acceptedFilepaths, str)
}

func (b *FilepathCondition) GetConditionValues() []string {
	return b.acceptedFilepaths
}
