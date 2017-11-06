package configure

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Stage struct {
	Script []string
	Env    []string
}

type RunConfig struct {
	Image        string
	Services     []string
	BeforeStages Stage `yaml:"before_stages"`
	AfterStages  Stage `yaml:"after_stages"`
	Build        Stage
	Test         Stage
	Deploy       Stage
}

func readConfig(filePath string) []byte {
	config_yml, err := ioutil.ReadFile(filePath)
	check(err)
	return config_yml

}

// Get build configuration of job by string of config
// unmarshals contents to RunConfig struct.
// returns struct and any error from unmarshaling
func GetRunConfig(filePath string) (RunConfig, error) {
	config_bytes := readConfig(filePath)
	runConf := RunConfig{}
	err := yaml.Unmarshal(config_bytes, &runConf)
	return runConf, err
}

// func main() {
//     run_config := getRunConfig("./test/test.yml")
//     fmt.Println(run_config)
// }