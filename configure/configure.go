package configure

import (
    "io/ioutil"
    "log"
    "gopkg.in/yaml.v2"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}


type Stage struct {
    Script []string
    Env []string
}

type RunConfig struct {
    Image string
    Services []string
    BeforeStages Stage `yaml:"before_stages"`
    AfterStages Stage  `yaml:"after_stages"`
    Build Stage
    Test Stage
    Deploy Stage

}

func readConfig(filePath string) []byte {
    config_yml, err := ioutil.ReadFile(filePath)
    check(err)
    return config_yml

}


func GetRunConfig(filePath string) RunConfig {
    config_bytes := readConfig(filePath)
    runConf := RunConfig{}
    err := yaml.Unmarshal(config_bytes, &runConf)
    if err != nil {
        log.Fatalf("error: %v", err)
    }
    return runConf
}


// func main() {
//     run_config := getRunConfig("./test/test.yml")
//     fmt.Println(run_config)    
// }