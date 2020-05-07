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
	loggingTime           map[string]time.Time
	currentInitiate       initiateData
}

func NewReinitiator(client *vici.ClientConn, logger log.Logger) *Reinitiator {
	initiateWorkerChannel := make(chan initiateData)
	go initiateWorker(log.With("type", "initiateWorker"), client, initiateWorkerChannel)
	return &Reinitiator{
		client:                client,
		logger:                logger,
		initiateWorkerChannel: initiateWorkerChannel,
		loggingTime:           map[string]time.Time{},
	}
}

func (i *Reinitiator) IKESAStatus(ikeSAStatus IKESAStatus) {
	for _, childSA := range ikeSAStatus.ChildSA {
		if childSA.State != nil {
			continue
		}
		initiate := initiateData{
			IKEName:   ikeSAStatus.Name,
			ChildName: childSA.Name,
		}
		select {
		case i.initiateWorkerChannel <- initiate:
			// initiation started
			i.currentInitiate = initiate
		default:
			if loggingTime, ok := i.loggingTime[initiate.getFullName()]; !ok || time.Now().Sub(loggingTime) >= 30*time.Second {
				log.Infof("Skip initiating Child SA %s, because Child SA %s is being initiated", initiate.getFullName(), i.currentInitiate.getFullName())
				i.loggingTime[initiate.getFullName()] = time.Now()
			}
		}
	}
}

func initiateWorker(logger log.Logger, client *vici.ClientConn, workerChannel chan initiateData) {
	for {
		initiateData := <-workerChannel

		logger.Infof("Initiating a Child SA for %s", initiateData.getFullName())
		err := client.Initiate(initiateData.ChildName, initiateData.IKEName, func(fields map[string]interface{}) {
			msg, _ := fields["msg"]
			logger.With("strongswanFields", fields).Infof("Initiating log for %s: %s", initiateData.getFullName(), msg)
		})

		if err != nil {
			logger.Errorf("got error trying to initiate Child SA %s: %s", initiateData.getFullName(), err)
			continue
		}
		logger.Infof("Initiated new Child SA %s", initiateData.getFullName())
	}
}

type initiateData struct {
	IKEName   string
	ChildName string
}

func (i initiateData) getFullName() string {
	return fmt.Sprintf("%s.%s", i.IKEName, i.ChildName)
}
