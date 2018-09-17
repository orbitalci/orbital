package trigger

import (
	//"testing"
	"testing"
)

/*
- branch: master||develop|| release.* and text: schema_changed
	changelist:
	- filepaths:
		- src/main/java/com/o8/kontrol.java
	- text:
		- test1
		- i did a thing
		- [schema_changed]
	- branch: release/update_schema
- branch: fix.* and text: buildme and filepath: GisCommon
	changelist:
	- filepaths:
		- GisCommon/src/main/java/com/...
	- text:
		- i did a big thing
		- [buildme]
	- branch: fix/memory_leak
- branch: master||develop and filepath: src/test && src/main
	changelist:
	- filepaths:
		- src/main/java/com/hilhil.java
		- src/test/java/com/hilhil.java
	- branch: develop
	- text:
		- [sizzurp]
*/

func TestConditionalDirective_Passes(t *testing.T) {
	//td := &ChangesetData{
	//	filesChanged: []string{
	//		"src/main/java/com/o8/controller.java",
	//		"src/main/test/com/o8/controllerTest.java",
	//		"deploy/helm/upgrade.yaml",
	//		"sql/__SQL/__SQL_1.sql",
	//		"hummuna",
	//	},
	//	commitTexts: []string{
	//		"i changed a controller",
	//		"i updated the schema [schema_changed]",
	//		"i added a controller test",
	//		"i added to deployment",
	//	},
	//	branch: "fix/thisiskewl",
	//}
	//
}