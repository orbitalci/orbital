package valet

import (
	"errors"
	"sync"
	"time"

	"bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/models"
)

func NewKillaValet() *KillaValet {
	return &KillaValet{contexts:make(map[string]*models.BuildContext)}
}

type KillaValet struct {
	contexts map[string]*models.BuildContext
}

func (kv *KillaValet) ListenForKillRequests(hashKillChan chan string) {
	for {
		time.Sleep(time.Millisecond)
		hash := <- hashKillChan
		kv.Kill(hash)
	}
}


func (kv *KillaValet) Kill(killHash string) error {
	ctx, active := kv.contexts[killHash]
	if !active {
		log.Log().Warning("hash was already complete, ", killHash)
		return errors.New("hash " + killHash + " was already complete")
	}
	ctx.CancelFunc()
	return nil
}

func (kv *KillaValet) ListenBuilds(buildsChan chan *models.BuildContext, mapLock sync.Mutex) {
	for newBuild := range buildsChan {
		mapLock.Lock()
		log.Log().Debug("got new build context for ", newBuild.Hash)
		kv.contexts[newBuild.Hash] = newBuild
		mapLock.Unlock()

	}
}


func (kv *KillaValet) contextCleanup(buildCtx *models.BuildContext, mapLock sync.Mutex) {
	select {
	case <-buildCtx.Context.Done():
		log.Log().Debugf("build for hash %s is complete", buildCtx.Hash)
		mapLock.Lock()
		defer mapLock.Unlock()
		if _, ok := kv.contexts[buildCtx.Hash]; ok {
			delete(kv.contexts, buildCtx.Hash)
		}
		// should this be unlock?
		//mapLock.Lock()
		return
	}
}