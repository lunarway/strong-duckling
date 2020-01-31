package vici

import (
	"fmt"
)

type Connection struct {
	ConnConf map[string]IKEConf `json:"connections"`
}

/*
Mapped from man configuration documentation
https://manpages.debian.org/testing/strongswan-swanctl/swanctl.conf.5.en.html#SETTINGS
*/

type IKEConf struct {
	// IKEVersion is the IKE protocol version: 1 for IKEv1, 2 for IKEv2 and 0 to
	// accept both.
	IKEVersion      string   `json:"version"`
	LocalAddresses  []string `json:"local_addrs"`
	RemoteAddresses []string `json:"remote_addrs,omitempty"`
	LocalPort       string   `json:"local_port,omitempty"`
	RemotePort      string   `json:"remote_port,omitempty"`
	Proposals       []string `json:"proposals,omitempty"`
	// VIPs are virtual IPs to use.
	VIPs []string `json:"vips,omitempty"`
	// Aggressive indicates if Aggressive Mode is used instead of Main mode fir
	// Identity Protection.
	Aggressive string `json:"aggressive"`
	// Pull indicates if Mode Config works in pull mode. If "no" push mode is
	// used.
	Pull string `json:"pull"`
	// DSCP is the differentiated services field codepoint set on outgoing IKE
	// packets. Value is a six digit binary encoded string referencing RFC 2474.
	DSCP string `json:"dscp"`
	// Encapsulation indicates if UDP encapsulation of ESP packets is enables.
	Encapsulation string `json:"encap"`
	MOBIKE        string `json:"mobike,omitempty"`
	// ReauthTimeSeconds is the re-authentication interval in seconds.
	ReauthTimeSeconds string `json:"reauth_time,omitempty"`
	// RekeyTimeSeconds is the rekeying interval in seconds.
	RekeyTimeSeconds string `json:"rekey_time"`
	// DPDDelay is the interval on which to check liveness of a peer. Is only
	// enforced if no IKE or ESP/AH packet has been received for the delay.
	DPDDelay string `json:"dpd_delay,omitempty"`
	// DPDTimeout specifies a custom interval for liveness of a peer in IKEv1.
	DPDTimeout string `json:"dpd_timeout,omitempty"`
	// Fragmentation controls if oversized IKE messages will be sent in fragments.
	// Possible values are yes (default), accept, force and no.
	Fragmentation   string `json:"fragmentation,omitempty"`
	Childless       string `json:"childless,omitempty"`
	SendCertRequest string `json:"send_certreq,omitempty"`
	SendCert        string `json:"send_cert,omitempty"`
	// PPKID identifies the Postquantum Preshared Key to use.
	PPKID       string `json:"ppk_id,omitempty"`
	PPKRequired string `json:"ppk_required,omitempty"`
	KeyingTries string `json:"keyingtries,omitempty"`
	// Unique indicates the uniqueness policy used for the connection.
	Unique string `json:"unique,omitempty"`
	// OverTime is the hard IKE_SA lifetime in percentage of the longer of
	// RekeyTimeSeconds and ReauthTimeSeconds.
	OverTime string `json:"over_time,omitempty"`
	// RandTime is the time range from which to choose a random jitter value to
	// subtract from rekey/reauth times.
	RandTime string   `json:"rand_time,omitempty"`
	Pools    []string `json:"pools,omitempty"`
	// XFRMInterfaceIDIn is the XFRM interface if set on inbound policies/SA.
	XFRMInterfaceIDIn string `json:"if_id_in,omitempty"`
	// XFRMInterfaceIDIn is the XFRM interface if set on outbound policies/SA.
	XFRMInterfaceIDOut string              `json:"if_id_out,omitempty"`
	Mediation          string              `json:"mediation,omitempty"`
	MediatedBy         string              `json:"mediated_by,omitempty"`
	MediationPeer      string              `json:"mediation_peer,omitempty"`
	LocalAuthSection   map[string]AuthConf `json:"-"`
	RemoteAuthSection  map[string]AuthConf `json:"-"`

	Children map[string]ChildSAConf `json:"children"`
}

type AuthConf struct {
	// Class is the authentication type.
	Class            string `json:"class"`
	EAPType          string `json:"eap-type"`
	EAPVendor        string `json:"eap-vendor"`
	XAuth            string `json:"xauth"`
	RevocationPolicy string `json:"revocation"`
	IKEIdentity      string `json:"id"`
	// AAAID is the AAA authentication backend identity
	AAAID string `json:"aaa_id"`
	// EAPID is the identity for authentication
	EAPID   string   `json:"eap_id"`
	XAuthID string   `json:"xauth_id"`
	Groups  []string `json:"groups,omitempty"`
	Certs   []string `json:"certs,omitempty"`
	CACerts []string `json:"cacerts,omitempty"`
}

type ChildSAConf struct {
	LocalTrafficSelectors  []string `json:"local-ts,omitempty"`
	RemoteTrafficSelectors []string `json:"remote-ts,omitempty"`
	ESPProposals           []string `json:"esp_proposals,omitempty"` //aes128-sha1_modp1024
	StartAction            string   `json:"start_action"`            //none,trap,start
	CloseAction            string   `json:"close_action"`
	ReqID                  string   `json:"reqid,omitempty"`
	RekeyTimeSeconds       string   `json:"rekey_time"`
	ReplayWindow           string   `json:"replay_window,omitempty"`
	IPsecMode              string   `json:"mode"`
	InstallPolicy          string   `json:"policies"`
	UpDown                 string   `json:"updown,omitempty"`
	Priority               string   `json:"priority,omitempty"`
	MarkIn                 string   `json:"mark_in,omitempty"`
	MarkOut                string   `json:"mark_out,omitempty"`
	DpdAction              string   `json:"dpd_action,omitempty"`
	LifeTime               string   `json:"life_time,omitempty"`
	RekeyBytes             string   `json:"rekey_by´tes,omitempty"`
	RekeyPackets           string   `json:"rekey_packets,omitempty"`
}

func (c *ClientConn) LoadConn(conn *map[string]IKEConf) error {
	requestMap := &map[string]interface{}{}

	err := ConvertToGeneral(conn, requestMap)
	if err != nil {
		return fmt.Errorf("convert to general: %w", err)
	}

	msg, err := c.Request("load-conn", *requestMap)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	if msg["success"] != "yes" {
		return fmt.Errorf("unsuccessful LoadConn: %v", msg["errmsg"])
	}

	return nil
}
