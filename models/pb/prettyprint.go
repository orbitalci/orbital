package pb

import (
	"bytes"
	"text/template"
)

//ChangesetData struct {
//FilesChanged         []string `protobuf:"bytes,1,rep,name=filesChanged" json:"filesChanged,omitempty"`
//CommitTexts          []string `protobuf:"bytes,2,rep,name=commitTexts" json:"commitTexts,omitempty"`
//Branch               string   `protobuf:"bytes,3,opt,name=branch" json:"branch,omitempty"`

func (c *ChangesetData) PrettyPrint() string {
	pretty := `Branch: {{.Branch}}
Commit Messages: 
{{- range .CommitTexts }}
  - {{ . }}
{{ end }}
Files Changed: 
{{- range .FilesChanged }}
  - {{ . }}
{{ end }}
`
	templ, err := template.New("settingsxml").Parse(pretty)
	if err != nil {
		return ""
	}
	var printedChangeset bytes.Buffer
	_ = templ.Execute(&printedChangeset, c)
	return printedChangeset.String()
}
//
//
//
//func (wt *WerkerTask) PrettyPrint() string {
//	pretty := `CheckoutHash: {{.CheckoutHash}}
//VcsType: {{.VcsType}}
//FullName: {{.FullName}}
//Id: {{.Id}}
//Branch: {{.Branch}]
//SignaledBy: {{.SignaledBy}}
//`
//}