package vici

type Version struct {
	Daemon  string `json:"daemon"`
	Version string `json:"version"`
	Sysname string `json:"sysname"`
	Release string `json:"release"`
	Machine string `json:"machine"`
}

func (c *ClientConn) Version() (*Version, error) {
	msg, err := c.Request("version", nil)
	if err != nil {
		return nil, err
	}
	out := &Version{}
	err = ConvertFromGeneral(msg, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
