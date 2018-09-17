package trigger

type ChangesetData struct {
	filesChanged []string
	commitTexts  []string
	branch       string
}
