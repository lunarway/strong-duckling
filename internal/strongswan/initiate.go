package strongswan

import (
	"fmt"
	"time"

	"github.com/lunarway/strong-duckling/internal/vici"
	"github.com/prometheus/common/log"
)

var _ IKESAStatusReceiver = &Reinitiator{}

type Reinitiator struct {
	client                *vici.ClientConn
	logger                log.Logger
	initiateWorkerChannel chan initiateData
}

func NewReinitiator(client *vici.ClientConn, logger log.Logger) *Reinitiator {
	return &Reinitiator{
		client:                client,
		logger:                logger,
		initiateWorkerChannel: nil,
	}
}

func (i *Reinitiator) IKESAStatus(ikeSAStatus IKESAStatus) {
	for _, childSA := range ikeSAStatus.ChildSA {
		if childSA.State == nil {
			if i.initiateWorkerChannel == nil {
				i.initiateWorkerChannel = make(chan initiateData)
				go initiateWorker(log.With("type", "initiateWorker"), i.client, i.initiateWorkerChannel)
			}

			i.initiateWorkerChannel <- initiateData{
				IKEName:   ikeSAStatus.Name,
				ChildName: childSA.Name,
			}
		}
	}
}

func initiateWorker(logger log.Logger, client *vici.ClientConn, workerChannel chan initiateData) {
	loggingTime := map[string]time.Time{}
	var initiateChannel <-chan error
	var currentInitiate initiateData

	for {
		select {
		case initiateData := <-workerChannel:
			if initiateChannel != nil {
				if loggingTime, ok := loggingTime[initiateData.getFullName()]; ok && loggingTime.After(time.Now().Add(30*time.Second)) {
					log.Infof("Skip initiating Child SA %s, because Child SA %s is being initiated", initiateData.getFullName(), currentInitiate.getFullName())
				}
				continue
			}

			logger.Infof("Initiating a Child SA for %s", initiateData.getFullName())
			currentInitiate = initiateData
			initiateChannel = initiate(client, initiateData, logger)

		case initiateErr := <-initiateChannel:
			if initiateErr != nil {
				logger.Errorf("got error trying to initiate Child SA %s: %s", currentInitiate.getFullName(), initiateErr)
			} else {
				logger.Infof("Initiated new Child SA %s", currentInitiate.getFullName())
			}
			initiateChannel = nil
		}

	}
}

func initiate(client *vici.ClientConn, initiate initiateData, logger log.Logger) <-chan error {
	r := make(chan error)
	go func() {
		defer close(r)
		err := client.Initiate(initiate.ChildName, initiate.IKEName, func(fields map[string]interface{}) {
			msg, _ := fields["msg"]
			logger.With("strongswanFields", fields).Infof("Initiating log for %s: %s", initiate.getFullName(), msg)
		})
		if err != nil {
			r <- err
			return
		}
		r <- nil
	}()
	return r
}

type initiateData struct {
	IKEName   string
	ChildName string
}

func (i initiateData) getFullName() string {
	return fmt.Sprintf("%s.%s", i.IKEName, i.ChildName)
}
