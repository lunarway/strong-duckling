package vici

type Uptime struct {
	Running string `json:"running"`
	Since   string `json:"since"`
}

type Active struct {
	Critical string `json:"critical"`
	High     string `json:"high"`
	Medium   string `json:"medium"`
	Low      string `json:"low"`
}

type Workers struct {
	Total  string `json:"total"`
	Idle   string `json:"idle"`
	Active Active `json:"active"`
}

type Queues struct {
	Critical string `json:"critical"`
	High     string `json:"high"`
	Medium   string `json:"medium"`
	Low      string `json:"low"`
}

type IkeSas struct {
	Total    string `json:"total"`
	HalfOpen string `json:"half-open"`
}

type MallInfo struct {
	Sbrk string `json:"sbrk"`
	Mmap string `json:"mmap"`
	Used string `json:"used"`
	Free string `json:"free"`
}

type Stats struct {
	Uptime    Uptime   `json:"uptime"`
	Workers   Workers  `json:"workers"`
	Queues    Queues   `json:"queues"`
	IkeSas    IkeSas   `json:"ikesas"`
	Scheduled string   `json:"scheduled"`
	Plugins   []string `json:"plugins"`
	MallInfo  MallInfo `json:"mallinfo"`
}

// Stats returns IKE daemon statistics and load information.
func (c *ClientConn) Stats() (Stats, error) {
	msg, err := c.Request("stats", nil)
	if err != nil {
		return Stats{}, err
	}
	var stats Stats
	err = ConvertFromGeneral(msg, &stats)
	if err != nil {
		return Stats{}, err
	}
	return stats, nil
}
