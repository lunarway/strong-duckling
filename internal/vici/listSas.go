package vici

import (
	"strconv"
)

// IkeSa is an IKE Security Associasion from a list-sa event.
type IkeSa struct {
	UniqueID   string `json:"uniqueid"` //called ike_id in terminate() argument.
	IKEVersion string `json:"version"`
	// State is the state of the IKE SA: ESTABLISHED
	State         string `json:"state"`
	LocalHost     string `json:"local-host"`
	LocalPort     string `json:"local-port"`
	LocalID       string `json:"local-id"`
	RemoteHost    string `json:"remote-host"`
	RemotePort    string `json:"remote-port"`
	RemoteID      string `json:"remote-id"`
	RemoteXAuthID string `json:"remote-xauth-id"` //client username
	RemoteEAPID   string `json:"remote-eap-id"`
	// Initiator indicates if this SA is the initiator.
	Initiator string `json:"initiator"`
	// InitiatorSPI contains a hex encoded initiator SPI / cookie
	InitiatorSPI string `json:"initiator-spi"`
	// ResponderSPI contains a hex encoded responder SPI / cookie
	ResponderSPI        string `json:"responder-spi"`
	EncryptionAlgorithm string `json:"encr-alg"`
	EncryptionKeySize   string `json:"encr-keysize"`
	IntegrityAlgorithm  string `json:"integ-alg"`
	IntegrityKeySize    string `json:"integ-keysize"`
	// PRFAlgorithm is the pseudo-random function used for keying material.
	PRFAlgorithm string `json:"prf-alg"`
	DHGroup      string `json:"dh-group"`
	// EstablishedSeconds is the number of seconds the IKE SA has been established.
	EstablishedSeconds string `json:"established"` // metric
	// RekeyTimeSeconds is the number of seconds before the IKE SA gets rekeyed.
	RekeyTimeSeconds string `json:"rekey-time"`
	// ReauthTimeSeconds is the number of seconds before the IIKE SA gets re-authenticated.
	ReauthTimeSeconds string `json:"reauth-time"`
	// LocalVIPs are the virtual IPs assigned by the remote peer, installed locally.
	LocalVIPs []string `json:"local-vips"`
	// RemoteVIPs are the virtual IPs assigned to the remote peer.
	RemoteVIPs []string `json:"remote-vips"`
	// ChildSAs is a map of IKE Child SAs keyed by their name.
	ChildSAs map[string]ChildSA `json:"child-sas"`
	/*
		Unmapped fields
		tasks-queued = [
				<list of currently queued tasks for execution>
		]
		tasks-active = [
				<list of tasks currently initiating actively>
		]
		tasks-passive = [
				<list of tasks currently handling passively>
		]
	*/
}

type ChildSA struct {
	Name     string `json:"name"`
	UniqueID string `json:"uniqueid"`
	ReqID    string `json:"reqid"`
	// State is the IKE Child SA state: INSTALLED
	State string `json:"state"`
	// IPsecMode is the IPsec mode: tunnel, transport, beet
	IPsecMode string `json:"mode"`
	// IPsecProtocol is the IPsec protocol: AH, ESP
	IPsecProtocol string `json:"protocol"`
	// UDPEncapsulation is "yes" if UDP encapsulation is enabled.
	UDPEncapsulation string `json:"encap"`
	// SPIIn contains a hex encoded inbound SPI.
	SPIIn string `json:"spi-in"`
	// SPIOut contains a hex encoded outbound SPI.
	SPIOut string `json:"spi-out"`
	// CPIIn contains a hex encoded inbound CPI if compression is used.
	CPIIn string `json:"cpi-in"`
	// CPIOut contains a hex encoded outbound CPI if compression is used.
	CPIOut              string `json:"cpi-out"`
	EncryptionAlgorithm string `json:"encr-alg"`
	EncryptionKeySize   string `json:"encr-keysize"`
	IntegrityAlgorithm  string `json:"integ-alg"`
	IntegrityKeySize    string `json:"integ-keysize"`
	// PRFAlgorithm is the pseudo-random function used for keying material.
	PRFAlgorithm string `json:"prf-alg"`
	DHGroup      string `json:"dh-group"`
	// ExtendedSequenceNumber indicates whether the SA is using extended sequence
	// numbers. If the value is 1 it is used otherwise it is empty.
	ExtendedSequenceNumber string `json:"esn"`
	BytesIn                string `json:"bytes-in"`
	BytesOut               string `json:"bytes-out"`
	PacketsIn              string `json:"packets-in"`
	PacketsOut             string `json:"packets-out"`
	// LastPacketInSeconds is the number of seconds since the last received packet.
	LastPacketInSeconds string `json:"use-in"`
	// LastPacketOutSeconds is the number of seconds since the last transmitted packet.
	LastPacketOutSeconds string `json:"use-out"`
	// RekeyTimeSeconds is the number of seconds before the IKE Child SA gets rekeyed.
	RekeyTimeSeconds string `json:"rekey-time"`
	// LifeTimeSeconds is the number of seconds before the IKE Child SA expires.
	LifeTimeSeconds string `json:"life-time"`
	// InstallTimeSeconds is the number of seconds the IKE Child SA has been installed.
	InstallTimeSeconds     string   `json:"install-time"`
	LocalTrafficSelectors  []string `json:"local-ts"`
	RemoteTrafficSelectors []string `json:"remote-ts"`
	/*
		Unmapped fields
		mark-in = <hex encoded inbound Netfilter mark value>
		mark-mask-in = <hex encoded inbound Netfilter mark mask>
		mark-out = <hex encoded outbound Netfilter mark value>
		mark-mask-out = <hex encoded outbound Netfilter mark mask>
		if-id-in = <hex encoded inbound XFRM interface ID>
		if-id-out = <hex encoded outbound XFRM interface ID>
	*/
}

func (s *ChildSA) GetBytesIn() uint64 {
	num, err := strconv.ParseUint(s.BytesIn, 10, 64)
	if err != nil {
		return 0
	}
	return num
}

func (s *ChildSA) GetBytesOut() uint64 {
	num, err := strconv.ParseUint(s.BytesOut, 10, 64)
	if err != nil {
		return 0
	}
	return num
}

// To be simple, list all clients that are connecting to this server .
// A client is a sa.
// Lists currently active IKE_SAs
func (c *ClientConn) ListSas(ike string, ike_id string) ([]map[string]IkeSa, error) {
	sas := []map[string]IkeSa{}
	var eventErr error
	//register event
	err := c.RegisterEvent("list-sa", func(response map[string]interface{}) {
		sa := &map[string]IkeSa{}
		err := convertFromGeneral(response, sa)
		if err != nil {
			eventErr = err
			return
		}
		sas = append(sas, *sa)
	})
	if err != nil {
		return nil, err
	}
	if eventErr != nil {
		return nil, eventErr
	}

	inMap := map[string]interface{}{}
	if ike != "" {
		inMap["ike"] = ike
	}
	if ike_id != "" {
		inMap["ike_id"] = ike_id
	}
	_, err = c.Request("list-sas", inMap)
	if err != nil {
		return nil, err
	}
	err = c.UnregisterEvent("list-sa")
	if err != nil {
		return nil, err
	}
	return sas, nil
}

//a vpn conn in the strongswan server
type VpnConnInfo struct {
	IkeSa IkeSa
	// FIXME: This looks wrong. JSON keys between IkeSa and ChildSAs are conflicting.
	ChildSA     ChildSA
	IkeSaName   string //looks like conn name in ipsec.conf, content is same as ChildSaName
	ChildSaName string //looks like conn name in ipsec.conf
}

func (c *VpnConnInfo) GuessUserName() string {
	if c.IkeSa.RemoteXAuthID != "" {
		return c.IkeSa.RemoteXAuthID
	}
	if c.IkeSa.RemoteID != "" {
		return c.IkeSa.RemoteID
	}
	return ""
}

// a helper method to avoid complex data struct in ListSas
// if it only have one child_sas ,it will put it into info.ChildSAs
func (c *ClientConn) ListAllVpnConnInfo() ([]VpnConnInfo, error) {
	sasList, err := c.ListSas("", "")
	if err != nil {
		return nil, err
	}
	list := make([]VpnConnInfo, len(sasList))
	for i, sa := range sasList {
		info := VpnConnInfo{}
		for ikeSaName, ikeSa := range sa {
			info.IkeSaName = ikeSaName
			info.IkeSa = ikeSa
			for childSaName, childSa := range ikeSa.ChildSAs {
				info.ChildSaName = childSaName
				info.ChildSA = childSa
				break
			}
			break
		}
		if len(info.IkeSa.ChildSAs) == 1 {
			info.IkeSa.ChildSAs = nil
		}
		list[i] = info
	}
	return list, nil
}
