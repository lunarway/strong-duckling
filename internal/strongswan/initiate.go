package strongswan

import (
	"github.com/lunarway/strong-duckling/internal/vici"
	"github.com/prometheus/common/log"
)

var _ IKESAStatusReceiver = &Reinitiator{}

type Reinitiator struct {
	client     *vici.ClientConn
	logger     log.Logger
	initiating bool
}

func NewReinitiator(client *vici.ClientConn, logger log.Logger) *Reinitiator {
	return &Reinitiator{
		client:     client,
		logger:     logger,
		initiating: false,
	}
}

func (i *Reinitiator) IKESAStatus(ikeSAStatus IKESAStatus) {
	for _, childSA := range ikeSAStatus.ChildSA {
		if childSA.State == nil && !i.initiating {
			log := log.
				With("ikeName", ikeSAStatus.Name).
				With("childName", childSA.Name)

			i.initiating = true
			go func() {
				defer func() { i.initiating = false }()
				log.Infof("Initiating a Child SA for %s.%s", ikeSAStatus.Name, childSA.Name)
				err := i.client.Initiate(childSA.Name, ikeSAStatus.Name, func(line string) {
					log.Infof("Log for %s.%s: %s", ikeSAStatus.Name, childSA.Name, line)
				})

				if err != nil {
					log.Errorf("got error trying to initiate Child SA %s.%s: %s", ikeSAStatus.Name, childSA.Name, err)
				} else {
					log.Infof("Initiated new Child SA %s.%s", ikeSAStatus.Name, childSA.Name)
				}

			}()
		}
	}
}
