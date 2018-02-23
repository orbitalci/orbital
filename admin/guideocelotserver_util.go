package admin

import (
	"bitbucket.org/level11consulting/ocelot/util/handler"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	ocenet "bitbucket.org/level11consulting/go-til/net"
)

//when new configurations are added to the config channel, create bitbucket client and webhooks
func SetupCredentials(gosss models.GuideOcelotServer, config *models.VCSCreds) error {
	gos := gosss.(*guideOcelotServer)
	//hehe right now we only have bitbucket
	switch config.Type {
	case "bitbucket":
		bitbucketClient := &ocenet.OAuthClient{}
		bitbucketClient.Setup(config)

		bbHandler := handler.GetBitbucketHandler(config, bitbucketClient)
		err := bbHandler.Walk()
		if err != nil {
			return err
		}
	}
	configPath := config.BuildCredPath(config.Type, config.AcctName)
	err := gos.RemoteConfig.AddCreds(configPath, config)
	return err
}

func SetupRepoCredentials(gosss models.GuideOcelotServer, config *models.RepoCreds) error {
	// todo: probably should do some kind of test f they are valid or not? is there a way to test these creds
	gos := gosss.(*guideOcelotServer)
	configPath := config.BuildCredPath(config.Type, config.AcctName)
	err := gos.RemoteConfig.AddCreds(configPath, config)
	return err
}