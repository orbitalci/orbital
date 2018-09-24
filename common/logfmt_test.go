package common

import (
	"testing"

	"github.com/go-test/deep"
)

func TestMaybeStrip(t *testing.T) {
	output := MaybeStrip([]byte(gradleGarbage), true)

	output = MaybeStrip([]byte(otherGarbage), true)
	if diff := deep.Equal(string(output), cleanedGarbage); diff != nil {
		t.Error(diff)
	}
	output = MaybeStrip([]byte("drinky BUILD | [5A[1m<[0;32;1m=====[0;39;1m--------> 43% "), false)
	if string(output) != "drinky BUILD | [5A[1m<[0;32;1m=====[0;39;1m--------> 43% " {
		t.Error("should not have tried to format anything")
	}
}
const cleanedGarbage = `BUILD | 17:59:47.410 [main] INFO org.jooq.codegen.JavaGenerator -                          
BUILD | 17:59:47.410 [main] INFO org.jooq.codegen.JavaGenerator - Removing excess files    
BUILD | 
BUILD | 
BUILD | 
BUILD | 
BUILD | 
BUILD | > Task :app:compileJava
BUILD | Note: /.ocelot/2d3bf51bfd99a9eccdfa23ebb3928d09401fedc8/app/src/main/java/com/level11/o8/folio/utils/AuthService.java uses unchecked or unsafe operations.
BUILD | Note: Recompile with -Xlint:unchecked for details.
`
const otherGarbage = `BUILD | 17:59:47.410 [main] INFO org.jooq.codegen.JavaGenerator -                          
BUILD | 17:59:47.410 [main] INFO org.jooq.codegen.JavaGenerator - Removing excess files    
BUILD | [0K
BUILD | [0K
BUILD | [0K
BUILD | [0K
BUILD | [0K
BUILD | [5A[1m<[0;32;1m=====[0;39;1m--------> 43% EXECUTING [1m 58s][m[38D[1B[1m> :store:generateStoreJooqSchemaSource[m[38D[1B> IDLE[6D[1B> IDLE[6D[1B> IDLE[6D[1B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 58s][m[38D[1B[1m> :store:compileJava[m[0K[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 58s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 58s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 58s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 58s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 58s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 58s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 58s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 59s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 59s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 59s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 59s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 59s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 59s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 59s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 59s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 59s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [1m 59s][m[38D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [2m 0s][m[0K[37D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [2m 0s][m[37D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [2m 0s][m[37D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m=====[0;39;1m--------> 44% EXECUTING [2m 0s][m[37D[1B[1m> :store:compileJava[m[20D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 0s][m[37D[1B[1m> :app:compileJava[m[0K[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 0s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 0s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 0s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 0s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 0s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 1s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 1s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 1s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 1s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 1s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 1s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 1s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 1s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 1s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 1s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 2s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 2s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 2s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 2s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 2s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 2s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 2s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 2s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 2s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 2s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 3s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 3s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 3s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 3s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 3s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 3s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 3s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 3s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 3s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 3s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 4s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 4s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 4s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 4s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 4s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 4s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[1m<[0;32;1m======[0;39;1m-------> 50% EXECUTING [2m 4s][m[37D[1B[1m> :app:compileJava[m[18D[4B[5A[0K
BUILD | [1m> Task :app:compileJava[m[0K
BUILD | Note: /.ocelot/2d3bf51bfd99a9eccdfa23ebb3928d09401fedc8/app/src/main/java/com/level11/o8/folio/utils/AuthService.java uses unchecked or unsafe operations.[0K
BUILD | Note: Recompile with -Xlint:unchecked for details.[0K
`
const gradleGarbage = `BUILD | > Starting Daemon> Starting Daemon> Starting Daemon> Starting Daemon> Starting Daemon> Starting Daemon> Starting Daemon> Starting Daemon> Starting Daemon> Starting Daemon> Starting Daemon> IDLE<-------------> 0% INITIALIZING [0s]<-------------> 0% INITIALIZING [0s]> Evaluating settings<-------------> 0% INITIALIZING [0s]> Evaluating settings<-------------> 0% INITIALIZING [0s]> Evaluating settings<-------------> 0% INITIALIZING [0s]> Evaluating settings<-------------> 0% INITIALIZING [0s]> Evaluating settings<-------------> 0% INITIALIZING [0s]> Evaluating settings > Compiling /.ocelot/edcd2a02eaf73e3813192dfb0726027f989f8f7d/settings.gradle into local compilation cache > Compiling settings file '/.ocelot/edcd2a02eaf73e3813192dfb0726027f989f8f7d/settings.gradle' to cross build script cache<-------------> 0% INITIALIZING [0s]> Evaluating settings > Compiling /.ocelot/edcd2a02eaf73e3813192dfb0726027f989f8f7d/settings.gradle into local compilation cache > Compiling settings file '/.ocelot/edcd2a02eaf73e3813192dfb0726027f989f8f7d/settings.gradle' to cross build script cache<-------------> 0% INITIALIZING [0s]> Evaluating settings > Compiling /.ocelot/edcd2a02eaf73e3813192dfb0726027f989f8f7d/settings.gradle into local compilation cache > Compiling settings file '/.ocelot/edcd2a02eaf73e3813192dfb0726027f989f8f7d/settings.gradle' to cross build script cache<-------------> 0% INITIALIZING [0s]> Evaluating settings > Compiling /.ocelot/edcd2a02eaf73e3813192dfb0726027f989f8f7d/settings.gradle into local compilation cache > Compiling settings file '/.ocelot/edcd2a02eaf73e3813192dfb0726027f989f8f7d/settings.gradle' to cross build script cache<-------------> 0% INITIALIZING [1s]> Evaluating settings<-------------> 0% INITIALIZING [1s]> Evaluating settings<-------------> 0% CONFIGURING [1s]> Loading projects<-------------> 0% CONFIGURING [1s]> root project<-------------> 0% CONFIGURING [1s]> root project<-------------> 0% CONFIGURING [1s]> root project<-------------> 0% CONFIGURING [1s]> root project<-------------> 0% CONFIGURING [1s]> root project<-------------> 0% CONFIGURING [1s]> root project<-------------> 0% CONFIGURING [1s]> root project<-------------> 0% CONFIGURING [2s]> root project<-------------> 0% CONFIGURING [2s]> root project<-------------> 0% CONFIGURING [2s]> root project<-------------> 0% CONFIGURING [2s]> root project<-------------> 0% CONFIGURING [2s]> root project<-------------> 0% CONFIGURING [2s]> root project > Compiling /.ocelot/edcd2a02eaf73e3813192dfb0726027f989f8f7d/build.gradle into local compilation cache<-------------> 0% CONFIGURING [2s]> root project<-------------> 0% CONFIGURING [2s]> root project<===----------> 25% CONFIGURING [2s]> :app > Resolve dependencies of detachedConfiguration1<===----------> 25% CONFIGURING [2s]> :app > Resolve dependencies of detachedConfiguration1<===----------> 25% CONFIGURING [3s]> :app > Resolve dependencies of detachedConfiguration1<===----------> 25% CONFIGURING [3s]> :app > Resolve dependencies of detachedConfiguration1<===----------> 25% CONFIGURING [3s]> :app > Resolve dependencies of detachedConfiguration1 > org.springframework.boot.gradle.plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [3s]> :app > Resolve dependencies of detachedConfiguration1 > org.springframework.boot.gradle.plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [3s]> :app > Resolve dependencies of detachedConfiguration1 > org.springframework.boot.gradle.plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [3s]> :app > Resolve dependencies of detachedConfiguration1 > org.springframework.boot.gradle.plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [3s]> :app > Resolve dependencies of detachedConfiguration1 > org.springframework.boot.gradle.plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [3s]> :app > Resolve dependencies of detachedConfiguration1 > org.springframework.boot.gradle.plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [3s]> :app > Resolve dependencies of detachedConfiguration1 > org.springframework.boot.gradle.plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [3s]> :app > Resolve dependencies of detachedConfiguration1 > org.springframework.boot.gradle.plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [4s]> :app > Resolve dependencies of detachedConfiguration1 > org.springframework.boot.gradle.plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [4s]> :app > Resolve dependencies of detachedConfiguration1 > org.springframework.boot.gradle.plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [4s]> :app > Resolve dependencies of detachedConfiguration1<===----------> 25% CONFIGURING [4s]> :app > Resolve dependencies of detachedConfiguration1<===----------> 25% CONFIGURING [4s]> :app > Resolve dependencies of detachedConfiguration1<===----------> 25% CONFIGURING [4s]> :app > Resolve dependencies of detachedConfiguration1<===----------> 25% CONFIGURING [4s]> :app > Resolve dependencies of detachedConfiguration1<===----------> 25% CONFIGURING [4s]> :app > Resolve dependencies of detachedConfiguration1<===----------> 25% CONFIGURING [4s]> :app > Resolve dependencies of detachedConfiguration1<===----------> 25% CONFIGURING [4s]> :app > Resolve dependencies of detachedConfiguration1<===----------> 25% CONFIGURING [5s]> :app > Resolve dependencies of detachedConfiguration2 > com.palantir.docker.gradle.plugin-0.17.2.pom<===----------> 25% CONFIGURING [5s]> :app > Resolve dependencies of detachedConfiguration2 > com.palantir.docker.gradle.plugin-0.17.2.pom<===----------> 25% CONFIGURING [5s]> :app > Resolve dependencies of detachedConfiguration2 > com.palantir.docker.gradle.plugin-0.17.2.pom<===----------> 25% CONFIGURING [5s]> :app > Resolve dependencies of detachedConfiguration2 > com.palantir.docker.gradle.plugin-0.17.2.pom<===----------> 25% CONFIGURING [5s]> :app > Resolve dependencies of detachedConfiguration2 > com.palantir.docker.gradle.plugin-0.17.2.pom<===----------> 25% CONFIGURING [5s]> :app > Resolve dependencies of detachedConfiguration2 > com.palantir.docker.gradle.plugin-0.17.2.pom<===----------> 25% CONFIGURING [5s]> :app > Resolve dependencies of detachedConfiguration2 > com.palantir.docker.gradle.plugin-0.17.2.pom<===----------> 25% CONFIGURING [5s]> :app > Resolve dependencies of detachedConfiguration2 > com.palantir.docker.gradle.plugin-0.17.2.pom<===----------> 25% CONFIGURING [5s]> :app > Resolve dependencies of detachedConfiguration2 > com.palantir.docker.gradle.plugin-0.17.2.pom<===----------> 25% CONFIGURING [5s]> :app > Resolve dependencies of detachedConfiguration2 > com.palantir.docker.gradle.plugin-0.17.2.pom<===----------> 25% CONFIGURING [6s]> :app > Resolve dependencies of detachedConfiguration2 > com.palantir.docker.gradle.plugin-0.17.2.pom<===----------> 25% CONFIGURING [6s]> :app > Resolve dependencies of detachedConfiguration2 > com.palantir.docker.gradle.plugin-0.17.2.pom<===----------> 25% CONFIGURING [6s]> :app > Resolve dependencies of detachedConfiguration2<===----------> 25% CONFIGURING [6s]> :app > Resolve dependencies of detachedConfiguration2<===----------> 25% CONFIGURING [6s]> :app > Resolve dependencies of detachedConfiguration2<===----------> 25% CONFIGURING [6s]> :app > Resolve dependencies of detachedConfiguration2<===----------> 25% CONFIGURING [6s]> :app > Resolve dependencies of detachedConfiguration2<===----------> 25% CONFIGURING [6s]> :app > Resolve dependencies of detachedConfiguration2<===----------> 25% CONFIGURING [6s]> :app > Resolve dependencies of detachedConfiguration2<===----------> 25% CONFIGURING [6s]> :app > Resolve dependencies of detachedConfiguration2<===----------> 25% CONFIGURING [7s]> :app > Resolve dependencies of detachedConfiguration2<===----------> 25% CONFIGURING [7s]> :app > Resolve dependencies of detachedConfiguration2<===----------> 25% CONFIGURING [7s]> :app > Resolve dependencies of detachedConfiguration2<===----------> 25% CONFIGURING [7s]> :app > Resolve dependencies of detachedConfiguration2<===----------> 25% CONFIGURING [7s]> :app > Resolve dependencies of :app:classpath > spring-boot-gradle-plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [7s]> :app > Resolve dependencies of :app:classpath > spring-boot-gradle-plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [7s]> :app > Resolve dependencies of :app:classpath > spring-boot-gradle-plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [7s]> :app > Resolve dependencies of :app:classpath > spring-boot-gradle-plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [7s]> :app > Resolve dependencies of :app:classpath > spring-boot-gradle-plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [7s]> :app > Resolve dependencies of :app:classpath > spring-boot-gradle-plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [8s]> :app > Resolve dependencies of :app:classpath > spring-boot-gradle-plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [8s]> :app > Resolve dependencies of :app:classpath > spring-boot-gradle-plugin-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [8s]> :app > Resolve dependencies of :app:classpath > spring-boot-tools-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [8s]> :app > Resolve dependencies of :app:classpath > spring-boot-tools-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [8s]> :app > Resolve dependencies of :app:classpath > spring-boot-tools-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [8s]> :app > Resolve dependencies of :app:classpath > spring-boot-tools-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [8s]> :app > Resolve dependencies of :app:classpath > spring-boot-tools-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [8s]> :app > Resolve dependencies of :app:classpath > spring-boot-tools-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [8s]> :app > Resolve dependencies of :app:classpath > spring-boot-parent-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [8s]> :app > Resolve dependencies of :app:classpath > spring-boot-parent-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [9s]> :app > Resolve dependencies of :app:classpath > spring-boot-parent-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [9s]> :app > Resolve dependencies of :app:classpath > spring-boot-parent-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [9s]> :app > Resolve dependencies of :app:classpath > spring-boot-parent-1.5.12.RELEASE.pom > 13 KB/26 KB downloaded<===----------> 25% CONFIGURING [9s]> :app > Resolve dependencies of :app:classpath > spring-boot-dependencies-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [9s]> :app > Resolve dependencies of :app:classpath > spring-boot-dependencies-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [9s]> :app > Resolve dependencies of :app:classpath > spring-boot-dependencies-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [9s]> :app > Resolve dependencies of :app:classpath > spring-boot-dependencies-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [9s]> :app > Resolve dependencies of :app:classpath > spring-boot-dependencies-1.5.12.RELEASE.pom<===----------> 25% CONFIGURING [9s]> :app > Resolve dependencies of :app:classpath > spring-boot-dependencies-1.5.12.RELEASE.pom > 40 KB/107 KB downloaded<===----------> 25% CONFIGURING [9s]> :app > Resolve dependencies of :app:classpath > jackson-bom-2.8.11.20180217.pom<===----------> 25% CONFIGURING [10s]> :app > Resolve dependencies of :app:classpath > jackson-bom-2.8.11.20180217.pom<===----------> 25% CONFIGURING [10s]> :app > Resolve dependencies of :app:classpath > jackson-bom-2.8.11.20180217.pom<===----------> 25% CONFIGURING [10s]> :app > Resolve dependencies of :app:classpath > jackson-bom-2.8.11.20180217.pom<===----------> 25% CONFIGURING [10s]> :app > Resolve dependencies of :app:classpath > jackson-parent-2.8.pom<===----------> 25% CONFIGURING [10s]> :app > Resolve dependencies of :app:classpath > jackson-parent-2.8.pom<===----------> 25% CONFIGURING [10s]> :app > Resolve dependencies of :app:classpath > jackson-parent-2.8.pom<===----------> 25% CONFIGURING [10s]> :app > Resolve dependencies of :app:classpath > jackson-parent-2.8.pom<===----------> 25% CONFIGURING [10s]> :app > Resolve dependencies of :app:classpath > oss-parent-27.pom<===----------> 25% CONFIGURING [10s]> :app > Resolve dependencies of :app:classpath > oss-parent-27.pom<===----------> 25% CONFIGURING [10s]> :app > Resolve dependencies of :app:classpath > oss-parent-27.pom<===----------> 25% CONFIGURING [11s]> :app > Resolve dependencies of :app:classpath > oss-parent-27.pom<===----------> 25% CONFIGURING [11s]> :app > Resolve dependencies of :app:classpath > oss-parent-27.pom > 13 KB/19 KB downloaded<===----------> 25% CONFIGURING [11s]> :app > Resolve dependencies of :app:classpath > log4j-bom-2.7.pom<===----------> 25% CONFIGURING [11s]> :app > Resolve dependencies of :app:classpath > log4j-bom-2.7.pom<===----------> 25% CONFIGURING [11s]> :app > Resolve dependencies of :app:classpath > log4j-bom-2.7.pom<===----------> 25% CONFIGURING [11s]> :app > Resolve dependencies of :app:classpath > log4j-bom-2.7.pom<===----------> 25% CONFIGURING [11s]> :app > Resolve dependencies of :app:classpath > apache-9.pom<===----------> 25% CONFIGURING [11s]> :app > Resolve dependencies of :app:classpath > apache-9.pom<===----------> 25% CONFIGURING [11s]> :app > Resolve dependencies of :app:classpath > apache-9.pom<===----------> 25% CONFIGURING [11s]> :app > Resolve dependencies of :app:classpath > apache-9.pom<===----------> 25% CONFIGURING [12s]> :app > Resolve dependencies of :app:classpath<===----------> 25% CONFIGURING [12s]> :app > Resolve dependencies of :app:classpath > spring-framework-bom-4.3.16.RELEASE.pom<===----------> 25% CONFIGURING [12s]> :app > Resolve dependencies of :app:classpath > spring-framework-bom-4.3.16.RELEASE.pom<===----------> 25% CONFIGURING [12s]> :app > Resolve dependencies of :app:classpath > spring-framework-bom-4.3.16.RELEASE.pom<===----------> 25% CONFIGURING [12s]> :app > Resolve dependencies of :app:classpath > spring-data-releasetrain-Ingalls-SR11.pom<===----------> 25% CONFIGURING [12s]> :app > Resolve dependencies of :app:classpath > spring-data-releasetrain-Ingalls-SR11.pom<===----------> 25% CONFIGURING [12s]> :app > Resolve dependencies of :app:classpath > spring-data-releasetrain-Ingalls-SR11.pom<===----------> 25% CONFIGURING [12s]> :app > Resolve dependencies of :app:classpath > spring-data-releasetrain-Ingalls-SR11.pom<===----------> 25% CONFIGURING [12s]> :app > Resolve dependencies of :app:classpath > spring-data-build-1.9.11.RELEASE.pom<===----------> 25% CONFIGURING [12s]> :app > Resolve dependencies of :app:classpath > spring-data-build-1.9.11.RELEASE.pom<===----------> 25% CONFIGURING [13s]> :app > Resolve dependencies of :app:classpath > spring-data-build-1.9.11.RELEASE.pom<===----------> 25% CONFIGURING [13s]> :app > Resolve dependencies of :app:classpath > spring-data-build-1.9.11.RELEASE.pom<===----------> 25% CONFIGURING [13s]> :app > Resolve dependencies of :app:classpath > spring-integration-bom-4.3.15.RELEASE.pom<===----------> 25% CONFIGURING [13s]> :app > Resolve dependencies of :app:classpath > spring-integration-bom-4.3.15.RELEASE.pom<===----------> 25% CONFIGURING [13s]> :app > Resolve dependencies of :app:classpath > spring-integration-bom-4.3.15.RELEASE.pom<===----------> 25% CONFIGURING [13s]> :app > Resolve dependencies of :app:classpath > spring-integration-bom-4.3.15.RELEASE.pom<===----------> 25% CONFIGURING [13s]> :app > Resolve dependencies of :app:classpath > spring-integration-bom-4.3.15.RELEASE.pom<===----------> 25% CONFIGURING [13s]> :app > Resolve dependencies of :app:classpath > spring-security-bom-4.2.5.RELEASE.pom<===----------> 25% CONFIGURING [13s]> :app > Resolve dependencies of :app:classpath > spring-security-bom-4.2.5.RELEASE.pom<===----------> 25% CONFIGURING [13s]> :app > Resolve dependencies of :app:classpath > spring-security-bom-4.2.5.RELEASE.pom<===----------> 25% CONFIGURING [14s]> :app > Resolve dependencies of :app:classpath > spring-security-bom-4.2.5.RELEASE.pom<===----------> 25% CONFIGURING [14s]> :app > Resolve dependencies of :app:classpath > spring-security-bom-4.2.5.RELEASE.pom<===----------> 25% CONFIGURING [14s]> :app > Resolve dependencies of :app:classpath<===----------> 25% CONFIGURING [14s]> :app > Resolve dependencies of :app:classpath<===----------> 25% CONFIGURING [14s]> :app > Resolve dependencies of :app:classpath > gradle-docker-0.17.2.pom<===----------> 25% CONFIGURING [14s]> :app > Resolve dependencies of :app:classpath > gradle-docker-0.17.2.pom<===----------> 25% CONFIGURING [14s]> :app > Resolve dependencies of :app:classpath > gradle-docker-0.17.2.pom<===----------> 25% CONFIGURING [14s]> :app > Resolve dependencies of :app:classpath > gradle-docker-0.17.2.pom<===----------> 25% CONFIGURING [14s]> :app > Resolve dependencies of :app:classpath > gradle-docker-0.17.2.pom<===----------> 25% CONFIGURING [14s]> :app > Resolve dependencies of :app:classpath > gradle-docker-0.17.2.pom
`