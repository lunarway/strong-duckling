package vici

import (
	"fmt"
	"time"
)

const (
	EVENT_IKE_UPDOWN   = "ike-updown"
	EVENT_IKE_REKEY    = "ike-rekey"
	EVENT_CHILD_UPDOWN = "child-updown"
	EVENT_CHILD_REKEY  = "child-rekey"
)

type EventIkeSAUpDown struct {
	ChildSAs            map[string]*EventChildSAUpDown `json:"child-sas"`
	DHGroup             string                         `json:"dh-group"`
	EncryptionKeySize   string                         `json:"encr-keysize"`
	EncryptionAlgorithm string                         `json:"encr-alg"`
	EstablishedSeconds  string                         `json:"established"`
	InitiatorSPI        string                         `json:"initiator-spi"`
	IntegrityAlgorithm  string                         `json:"integ-alg"`
	LocalID             string                         `json:"local-id"`
	LocalHost           string                         `json:"local-host"`
	LocalPort           string                         `json:"local-port"`
	Nat_any             string                         `json:"nat-any"`
	Nat_remote          string                         `json:"nat-remote"`
	PRFAlgorithm        string                         `json:"prf-alg"`
	RekeyTimeSeconds    string                         `json:"rekey-time"`
	RemoteID            string                         `json:"remote-id"`
	RemoteHost          string                         `json:"remote-host"`
	RemotePort          string                         `json:"remote-port"`
	ResponderSPI        string                         `json:"responder-spi"`
	State               string                         `json:"state"`
	Task_Active         []string                       `json:"tasks-active"`
	UniqueID            string                         `json:"uniqueid"`
	IKEVersion          string                         `json:"version"`
}

type EventChildSAUpDown struct {
	BytesIn                string   `json:"bytes-in"`
	BytesOut               string   `json:"bytes-out"`
	UDPEncapsulation       string   `json:"encap"`
	EncryptionAlgorithm    string   `json:"encr-alg"`
	EncryptionKeySize      string   `json:"encr-keysize"`
	IntegrityAlgorithm     string   `json:"integ-alg"`
	InstallTimeSeconds     string   `json:"install-time"`
	LifeTimeSeconds        string   `json:"life-time"`
	LocalTrafficSelectors  []string `json:"local-ts"`
	IPsecMode              string   `json:"mode"`
	Name                   string   `json:"name"`
	IPsecProtocol          string   `json:"protocol"`
	PacketsOut             string   `json:"packets-out"`
	PacketsIn              string   `json:"packets-in"`
	RekeyTimeSeconds       string   `json:"rekey-time"`
	RemoteTrafficSelectors []string `json:"remote-ts"`
	ReqID                  string   `json:"reqid"`
	SPIIn                  string   `json:"spi-in"`
	SPIOut                 string   `json:"spi-out"`
	State                  string   `json:"state"`
	UniqueId               string   `json:"uniqueid"`
}

type EventIkeRekeyPair struct {
	New EventIkeRekeySA `json:"new"`
	Old EventIkeRekeySA `json:"old"`
}

type EventIkeRekeySA struct {
	ChildSAs            map[string]*EventChildRekeyPair `json:"child-sas"`
	DHGroup             string                          `json:"dh-group"`
	EncryptionAlgorithm string                          `json:"encr-alg"`
	EncryptionKeySize   string                          `json:"encr-keysize"`
	EstablishedSeconds  string                          `json:"established"`
	InitiatorSPI        string                          `json:"initiator-spi"`
	IntegrityAlgorithm  string                          `json:"integ-alg"`
	LocalHost           string                          `json:"local-host"`
	LocalPort           string                          `json:"local-port"`
	LocalID             string                          `json:"local-id"`
	Nat_any             string                          `json:"nat-any"`
	Nat_remote          string                          `json:"nat-remote"`
	PRFAlgorithm        string                          `json:"prf-alg"`
	RekeyTimeSeconds    string                          `json:"rekey-time"`
	RemoteID            string                          `json:"remote-id"`
	RemoteHost          string                          `json:"remote-host"`
	RemotePort          string                          `json:"remote-port"`
	ResponderSPI        string                          `json:"responder-spi"`
	State               string                          `json:"state"`
	Task_Active         []string                        `json:"tasks-active"`
	Task_Passive        []string                        `json:"tasks-passive"`
	UniqueID            string                          `json:"uniqueid"`
	IKEVersion          string                          `json:"version"`
}

type EventChildRekeyPair struct {
	New EventChildRekeySA `json:"new"`
	Old EventChildRekeySA `json:"old"`
}

type EventChildRekeySA struct {
	BytesIn                string   `json:"bytes-in"`
	BytesOut               string   `json:"bytes-out"`
	UDPEncapsulation       string   `json:"encap"`
	EncryptionAlgorithm    string   `json:"encr-alg"`
	EncryptionKeySize      string   `json:"encr-keysize"`
	IntegrityAlgorithm     string   `json:"integ-alg"`
	InstallTimeSeconds     string   `json:"install-time"`
	LifeTimeSeconds        string   `json:"life-time"`
	LocalTrafficSelectors  []string `json:"local-ts"`
	IPsecMode              string   `json:"mode"`
	Name                   string   `json:"name"`
	PacketsIn              string   `json:"packets-in"`
	PacketsOut             string   `json:"packets-out"`
	IPsecProtocol          string   `json:"protocol"`
	RemoteTrafficSelectors []string `json:"remote-ts"`
	RekeyTimeSeconds       string   `json:"rekey-time"`
	ReqID                  string   `json:"reqid"`
	SPIIn                  string   `json:"spi-in"`
	SPIOut                 string   `json:"spi-out"`
	State                  string   `json:"state"`
	LastPacketInSeconds    string   `json:"use-in"`
	LastPacketOutSeconds   string   `json:"use-out"`
	UniqueId               string   `json:"uniqueid"`
}

type EventIkeUpDown struct {
	Up  bool
	Ike map[string]*EventIkeSAUpDown
}

type EventIkeRekey struct {
	Ike map[string]*EventIkeRekeyPair
}

type EventChildRekey struct {
	Ike map[string]*EventIkeRekeySA
}

type EventChildUpDown struct {
	Up  bool
	Ike map[string]*EventIkeSAUpDown
}

type EventIkeSa struct {
	IkeSa
	TasksActive []string `json:"tasks-active"`
}

type EventInfo struct {
	Up  bool
	Ike map[string]*EventIkeSa
}

type MonitorCallBack func(event string, info interface{})

func handleIkeUpDown(eventName string, callback MonitorCallBack, response map[string]interface{}) error {
	event := &EventIkeUpDown{}
	event.Ike = map[string]*EventIkeSAUpDown{}
	//we need to marshall all ikes manual because json uses connections names as key
	for name := range response {
		value := response[name]
		if name == "up" {
			event.Up = true
		} else {
			sa := &EventIkeSAUpDown{}
			err := ConvertFromGeneral(value, sa)
			if err != nil {
				return fmt.Errorf("convert from general: %w", err)
			}
			event.Ike[name] = sa
		}
	}
	callback(eventName, event)
	return nil
}

func handleIkeRekey(eventName string, callback MonitorCallBack, response map[string]interface{}) error {
	event := &EventIkeRekey{}
	event.Ike = map[string]*EventIkeRekeyPair{}
	//we need to marshall all ikes manual because json uses connections names as key
	for name := range response {
		value := response[name]
		sa := &EventIkeRekeyPair{}
		err := ConvertFromGeneral(value, sa)
		if err != nil {
			return fmt.Errorf("convert from general: %w", err)
		}
		event.Ike[name] = sa
	}
	callback(eventName, event)
	return nil
}

func handleChildUpDown(eventName string, callback MonitorCallBack, response map[string]interface{}) error {
	event := &EventChildUpDown{}
	event.Ike = map[string]*EventIkeSAUpDown{}
	//we need to marshall all ikes manual because json uses connections names as key
	for name := range response {
		value := response[name]
		if name == "up" {
			event.Up = true
		} else {
			sa := &EventIkeSAUpDown{}
			err := ConvertFromGeneral(value, sa)
			if err != nil {
				return fmt.Errorf("convert from general: %w", err)
			}
			event.Ike[name] = sa
		}
	}
	callback(eventName, event)
	return nil
}

func handleChildRekey(eventName string, callback MonitorCallBack, response map[string]interface{}) error {
	event := &EventChildRekey{}
	event.Ike = map[string]*EventIkeRekeySA{}
	//we need to marshall all ikes manual because json uses connections names as key
	for name := range response {
		value := response[name]
		sa := &EventIkeRekeySA{}
		err := ConvertFromGeneral(value, sa)
		if err != nil {
			return fmt.Errorf("convert from general: %w", err)
		}
		event.Ike[name] = sa
	}
	callback(eventName, event)
	return nil
}

func (c *ClientConn) MonitorSA(callback MonitorCallBack, watchdog time.Duration) (err error) {
	//register event
	err = c.RegisterEvent(EVENT_CHILD_UPDOWN, func(response map[string]interface{}) {
		err := handleChildUpDown(EVENT_CHILD_UPDOWN, callback, response)
		if err != nil {
			fmt.Printf("Failed to handle EVENT_CHILD_UPDOWN: %v\n", err)
		}
	})
	if err != nil {
		return err
	}
	err = c.RegisterEvent(EVENT_CHILD_REKEY, func(response map[string]interface{}) {
		err := handleChildRekey(EVENT_CHILD_REKEY, callback, response)
		if err != nil {
			fmt.Printf("Failed to handle EVENT_CHILD_REKEY: %v\n", err)
		}
	})
	if err != nil {
		return err
	}
	err = c.RegisterEvent(EVENT_IKE_UPDOWN, func(response map[string]interface{}) {
		err := handleIkeUpDown(EVENT_IKE_UPDOWN, callback, response)
		if err != nil {
			fmt.Printf("Failed to handle EVENT_IKE_UPDOWN: %v\n", err)
		}
	})
	if err != nil {
		return err
	}
	err = c.RegisterEvent(EVENT_IKE_REKEY, func(response map[string]interface{}) {
		err := handleIkeRekey(EVENT_IKE_REKEY, callback, response)
		if err != nil {
			fmt.Printf("Failed to handle EVENT_IKE_REKEY: %v\n", err)
		}
	})
	if err != nil {
		return err
	}

	for {
		time.Sleep(watchdog)
		//collect some daemon stats to see if connection is alive
		if _, err := c.Stats(); err != nil {
			return err
		}
	}
}
