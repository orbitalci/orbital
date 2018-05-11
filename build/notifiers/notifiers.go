package notifiers

import (
	"github.com/shankj3/ocelot/models/pb"
)

// Interface for notifying on the status of a build
type Notifier interface {
	RunIntegration(intCreds []pb.OcyCredder, status *pb.Status, notifications *pb.Notifications) error
	SubType() pb.SubCredType
	String() string
	// IsRelevant will check the build config and then return a true if this build requires this type of notification
	IsRelevant(wc *pb.BuildConfig) bool
}

