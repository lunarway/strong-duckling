package strongswan

import (
	"github.com/lunarway/strong-duckling/internal/vici"
)

type IKESAStatusReceiver interface {
	IKESAStatus(ikeSAStatus IKESAStatus)
}

type IKESAStatus struct {
	Name          string
	Configuration vici.IKEConf
	State         *vici.IkeSa
	ChildSA       []ChildSAStatus
}

type ChildSAStatus struct {
	Name          string
	Configuration vici.ChildSAConf
	State         *vici.ChildSA
}
