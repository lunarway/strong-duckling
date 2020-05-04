package strongswan

import (
	"github.com/lunarway/strong-duckling/internal/vici"
	"github.com/prometheus/common/log"
)

var _ IKESAStatusReceiver = &Reinitiator{}

type Reinitiator struct {
	client *vici.ClientConn
	logger log.Logger
}

func NewReinitiator(client *vici.ClientConn, logger log.Logger) *Reinitiator {
	return &Reinitiator{
		client: client,
		logger: logger,
	}
}

func (i *Reinitiator) initate() {

}

func (i *Reinitiator) IKESAStatus(ikeSAStatus IKESAStatus) {
	for _, childSA := range ikeSAStatus.ChildSA {
		if childSA.State == nil {
			log := log.
				With("ikeName", ikeSAStatus.Name).
				With("childName", childSA.Name)

			log.Infof("Initiating a Child SA for %s.%s", ikeSAStatus.Name, childSA.Name)
			err := i.client.Initiate(childSA.Name, ikeSAStatus.Name)

			if err != nil {
				log.Errorf("got error trying to initiate Child SA %s.%s: %s", ikeSAStatus.Name, childSA.Name, err)
				return
			}

			log.Infof("Initiated new Child SA %s.%s", ikeSAStatus.Name, childSA.Name)
		}
	}
}
