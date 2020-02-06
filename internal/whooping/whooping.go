package whooping

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/prometheus/common/log"
)

type Whooper struct {
	latency time.Duration
	drift   time.Duration
	open    bool
}

func (whooper *Whooper) RegisterListener(serveMux *http.ServeMux, listeningAddress string) {
	serveMux.HandleFunc("/whoop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "only post allowed", http.StatusBadRequest)
		}

		whoop := Whoop{}
		err := json.NewDecoder(r.Body).Decode(&whoop)

		if err != nil {
			log.Debugf("Got error trying to parse whoop: %+v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		if whoop.Message == "whoop" {
			r.Header.Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Whoop{
				From:      listeningAddress,
				Message:   "whoop whoop",
				Timestamp: time.Now(),
				RemoteStatus: WhoopRemoteStatus{
					Open:    whooper.open,
					Latency: whooper.latency,
					Drift:   whooper.drift,
				},
			})
		} else {
			log.Debugf("Got error trying to answer whoop message: %s", whoop.Message)
			http.Error(w, "can't understand body", http.StatusBadRequest)
		}
	})
}

func (whooper *Whooper) Whoop(endpoint string, listeningEndpoint string) {
	log.Debugf("Sending whoop to %s with address %s", endpoint, listeningEndpoint)
	fullEndpoint := endpoint + "/whoop"

	now := time.Now()
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(Whoop{
		From:      listeningEndpoint,
		Message:   "whoop",
		Timestamp: now,
		RemoteStatus: WhoopRemoteStatus{
			Open:    whooper.open,
			Latency: whooper.latency,
			Drift:   whooper.drift,
		},
	})

	func() {
		resp, err := http.Post(fullEndpoint, "application/json", buf)
		if err != nil {
			log.Debugf("Got error whooping %s. Error: %s", fullEndpoint, err)
			whooper.open = false
			return
		}

		whoop := Whoop{}
		err = json.NewDecoder(resp.Body).Decode(&whoop)

		if err != nil {
			whooper.open = false
			log.Errorf("Got error trying to parse back-whoop from %s. Error: %s", fullEndpoint, err)
			return
		}

		if whoop.Message == "whoop whoop" {
			whooper.open = true
			whooper.latency = time.Since(now)
			whooper.drift = now.Add(whooper.latency / 2).Sub(whoop.Timestamp)
			log.Debugf("Got whoop whoop from %s. RemoteStatus is open is: %v and latency: %v and drift is %s",
				fullEndpoint, whoop.RemoteStatus.Open, whoop.RemoteStatus.Latency, whoop.RemoteStatus.Drift)
		} else {
			whooper.open = false
			log.Debugf("Got unexpected back-whoop response from %s. Message: \"%s\"", fullEndpoint, whoop.Message)
		}
	}()

	log.Debugf("Connection is open: %v and latency is: %v and drifts is %s", whooper.open, whooper.latency, whooper.drift)
}
