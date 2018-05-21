## Non-docker builds 

Some builds can't run in a docker container, because some giant corporations are a controlling bunch and don't want to make your life easier and less stateful. 

In this most unfortunate of cases, you can use nodes that are tagged by using the `machineTag` field in the `ocelot.yml` instead of the `image` field. If you do this, the build will only be executed on a werker node that is running with that tag. 

For example: 
```yaml
machineTag: ios
buildTool: xcode
branches:
  - ALL
stages:
  - name: build
    script:
      - UNITY_VERSION= 
      - COMMIT_NUMBER=$(git rev-list HEAD | wc -l | xargs)
      - VERSION=$(git describe --tags --long)
      -  |
        /Applications/Unity$UNITY_VERSION/Unity.app/Contents/MacOS/Unity \
        -quit -batchmode -nolog -logFile -nographics -projectPath ${WORKSPACE} \
        -buildTarget ios \
        -targetPath ${WORKSPACE}/OceanSnapper \
        -executeMethod BuildTools.Build.iOSBuild \
        -bundleVersion ${VERSION} \
        -bundleVersionCode ${COMMIT_NUMBER} \
        -hockeyAppId yyyy \
        -hockeyAppSecret xxxxx \
        -hockeyAppPackage com.whyapple.zzzz \
        -development \
        -exportMethod enterprise \
        -provisioningProfile "dumb profile"
  - name: build ios
    script:
      - cd ./OceanSnapper
      - ./xcode.command qqqq user
``` 

The repository with this `ocelot.yml` file will only run on a machine tagged with `ios`