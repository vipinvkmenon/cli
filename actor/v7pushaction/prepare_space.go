package v7pushaction

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func (actor Actor) PrepareSpace(spaceGUID string, parsedManifest ParsedManifest) <-chan *PushEvent {
	pushEventStream := make(chan *PushEvent)

	go func() {
		log.Debug("starting apply manifest go routine")
		defer close(pushEventStream)

		pushEventStream <- &PushEvent{Event: ApplyManifest}

		manifestBytes, err := yaml.Marshal(parsedManifest)
		if err != nil {
			pushEventStream <- &PushEvent{Err: err}
			return
		}

		warnings, err := actor.V7Actor.SetSpaceManifest(spaceGUID, manifestBytes)

		pushEventStream <- &PushEvent{Event: ApplyManifestComplete, Err: err, Warnings: Warnings(warnings)}
	}()

	return pushEventStream
}
