package trigger

import (
	"strings"
	
	"github.com/shankj3/ocelot/models/pb"
)

type Section interface {
	GetTriggerType() TriggerType
	// PassesMuster should be going through the relevant changeset data that it is given and figuring out
	//  if its trigger type is fulfilled in that given set. For example, the Branch TriggerType will be making sure the active
	//  branch in the changeset data regex matches with at least one of the branches in its acceptable branches list
	PassesMuster(*pb.ChangesetData) bool
	// GetLogical will retrieve the type of logical and/or that should be used for each condition value given for this particular trigger type
	GetLogical() Conditional
	SetLogical(Conditional)
	// AddConditionValue should add to the list of values that PassesMuster will check against. The ConditionValue for example
	//  could be 'master', or 'develop', etc for branches
	AddConditionValue(string)
	// GetConditionValues returns all the conditions values that have been added for that section. E.g. []string{"master", "develop", "release\/.*"} for branch
	GetConditionValues() []string
}

type BranchCondition struct {
	acceptedBranches []string
	logical          Conditional
}

func (b *BranchCondition) GetTriggerType() TriggerType {
	return Branch
}

// PassesMuster will check to make sure the branch in the changeset data regex matches with at least one of
// the accepted branches in its list
func (b *BranchCondition) PassesMuster(td *pb.ChangesetData) bool {
	ok, _ := BranchRegexOk(td.Branch, b.acceptedBranches)
	return ok
}

func (b *BranchCondition) GetLogical() Conditional {
	return b.logical
}

func (b *BranchCondition) SetLogical(conditional Conditional) {
	b.logical = conditional
}

// AddConditionalValue will add to the list of branches that will be checked against in PassesMuster
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

// PassesMuster will make sure that the text supplied as conditional values is found in the changeset data.
//  This is done along two different paths:
//    If the GetLogical() is OR:
//      At least one of the supplied commit texts given in the conditional values must be found in the
//      commit messages in the changeset data
//    If the GetLogical() is AND:
//      Every supplied commit text given in the conditional values must be found in the commit messages
func (b *TextCondition) PassesMuster(td *pb.ChangesetData) bool {
	return changesPassMuster(b.logical, td.CommitTexts, b.acceptedTexts)
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

// PassesMuster will make sure that the filepaths supplied as conditional values are found in the changeset data.
//  This is done along two different paths:
//    If the GetLogical() is OR:
//      At least one of the supplied filepaths given in the conditional values added by AddConditionalValue() must be
//      in the list of changed files in the ChangesetData
//    If the GetLogical() is AND:
//      All supplied filepaths given in the conditional values added by AddConditionalValue() must be
//      in the list of changed files in the ChangesetData
func (b *FilepathCondition) PassesMuster(td *pb.ChangesetData) bool {
	return changesPassMuster(b.logical, td.FilesChanged, b.acceptedFilepaths)
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
