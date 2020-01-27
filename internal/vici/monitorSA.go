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
	Child_sas     map[string]*EventChildSAUpDown `json:"child-sas"`
	Dh_group      string                         `json:"dh-group"`
	Encr_keysize  string                         `json:"encr-keysize"`
	Encr_alg      string                         `json:"encr-alg"`
	Established   string                         `json:"established"`
	Initiator_spi string                         `json:"initiator-spi"`
	Integ_alg     string                         `json:"integ-alg"`
	Local_id      string                         `json:"local-id"`
	Local_host    string                         `json:"local-host"`
	Local_port    string                         `json:"local-port"`
	Nat_any       string                         `json:"nat-any"`
	Nat_remote    string                         `json:"nat-remote"`
	Prf_alg       string                         `json:"prf-alg"`
	Rekey_time    string                         `json:"rekey-time"`
	Remote_id     string                         `json:"remote-id"`
	Remote_host   string                         `json:"remote-host"`
	Remote_port   string                         `json:"remote-port"`
	Responder_spi string                         `json:"responder-spi"`
	State         string                         `json:"state"`
	Task_Active   []string                       `json:"tasks-active"`
	Uniqueid      string                         `json:"uniqueid"`
	Version       string                         `json:"version"`
}

type EventChildSAUpDown struct {
	Bytes_in     string   `json:"bytes-in"`
	Bytes_out    string   `json:"bytes-out"`
	Encap        string   `json:"encap"`
	Encr_alg     string   `json:"encr-alg"`
	Encr_keysize string   `json:"encr-keysize"`
	Integ_alg    string   `json:"integ-alg"`
	Install_time string   `json:"install-time"`
	Life_time    string   `json:"life-time"`
	Local_ts     []string `json:"local-ts"`
	Mode         string   `json:"mode"`
	Name         string   `json:"name"`
	Protocol     string   `json:"protocol"`
	Packets_out  string   `json:"packets-out"`
	Packets_in   string   `json:"packets-in"`
	Rekey_time   string   `json:"rekey-time"`
	Remote_ts    []string `json:"remote-ts"`
	Reqid        string   `json:"reqid"`
	Spi_in       string   `json:"spi-in"`
	Spi_out      string   `json:"spi-out"`
	State        string   `json:"state"`
	UniqueId     string   `json:"uniqueid"`
}

type EventIkeRekeyPair struct {
	New EventIkeRekeySA `json:"new"`
	Old EventIkeRekeySA `json:"old"`
}

type EventIkeRekeySA struct {
	Child_sas     map[string]*EventChildRekeyPair `json:"child-sas"`
	Dh_group      string                          `json:"dh-group"`
	Encr_alg      string                          `json:"encr-alg"`
	Encr_keysize  string                          `json:"encr-keysize"`
	Established   string                          `json:"established"`
	Initiator_spi string                          `json:"initiator-spi"`
	Integ_alg     string                          `json:"integ-alg"`
	Local_host    string                          `json:"local-host"`
	Local_port    string                          `json:"local-port"`
	Local_id      string                          `json:"local-id"`
	Nat_any       string                          `json:"nat-any"`
	Nat_remote    string                          `json:"nat-remote"`
	Prf_alg       string                          `json:"prf-alg"`
	Rekey_time    string                          `json:"rekey-time"`
	Remote_id     string                          `json:"remote-id"`
	Remote_host   string                          `json:"remote-host"`
	Remote_port   string                          `json:"remote-port"`
	Responder_spi string                          `json:"responder-spi"`
	State         string                          `json:"state"`
	Task_Active   []string                        `json:"tasks-active"`
	Task_Passive  []string                        `json:"tasks-passive"`
	Uniqueid      string                          `json:"uniqueid"`
	Version       string                          `json:"version"`
}

type EventChildRekeyPair struct {
	New EventChildRekeySA `json:"new"`
	Old EventChildRekeySA `json:"old"`
}

type EventChildRekeySA struct {
	Bytes_in     string   `json:"bytes-in"`
	Bytes_out    string   `json:"bytes-out"`
	Encap        string   `json:"encap"`
	Encr_alg     string   `json:"encr-alg"`
	Encr_keysize string   `json:"encr-keysize"`
	Integ_alg    string   `json:"integ-alg"`
	Install_time string   `json:"install-time"`
	Life_time    string   `json:"life-time"`
	Local_ts     []string `json:"local-ts"`
	Mode         string   `json:"mode"`
	Name         string   `json:"name"`
	Packets_in   string   `json:"packets-in"`
	Packets_out  string   `json:"packets-out"`
	Protocol     string   `json:"protocol"`
	Remote_ts    []string `json:"remote-ts"`
	Rekey_time   string   `json:"rekey-time"`
	Reqid        string   `json:"reqid"`
	Spi_in       string   `json:"spi-in"`
	Spi_out      string   `json:"spi-out"`
	State        string   `json:"state"`
	Use_in       string   `json:"use-in"`
	Use_out      string   `json:"use-out"`
	UniqueId     string   `json:"uniqueid"`
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
