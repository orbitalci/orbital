package minioconfig

/*
{
  "version": "9",
  "hosts": {
      "minio-l11": {
        "url": "https://minio.metaverse.l11.com",
        "accessKey": "A1WAZD98ABN",
        "secretKey": "2309ASDKJC93BALS78ZNB30",
        "api": "s3v4",
        "lookup": "auto"
    }
  }
}
*/

func getMinioObj() *minioConfig {
	return &minioConfig{
		Version: defaultVersion,
		Hosts: make(map[string]*minioConfigEntry),
	}
}

type minioConfigEntry struct {
	Url  	  string	`json:"url"`
	AccessKey string	`json:"accessKey"`
	SecretKey string	`json:"secretKey"`
	Api 	  string	`json:"api"`
	Lookup    string	`json:"lookup"`
}

type minioConfig struct {
	Hosts 	map[string]*minioConfigEntry `json:"hosts"`
	Version string `json:"version,omitempty"`
}

const (
	defaultApi     = "s3v4"
	defaultLookup  = "auto"
	defaultVersion = "9"
)
