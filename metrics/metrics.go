package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"sort"

	"github.com/mtfelian/elixir-testnet-updater/notifier"
)

// Metrics represents metrics
type Metrics struct {
	uri         string
	notifier    notifier.Notifier
	lastMetrics map[string]any
}

// Params represents metrics parameters
type Params struct {
	URI      string
	Notifier notifier.Notifier
}

// New creates new metrics fetcher
func New(p Params) *Metrics {
	return &Metrics{
		uri:      p.URI,
		notifier: p.Notifier,
	}
}

// Fetch from the container's endpoint
func (m *Metrics) Fetch() (map[string]any, error) {
	const metricsURI = "/health"
	endpoint := fmt.Sprintf("%s%s", m.uri, metricsURI)
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error fetching metrics: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var metrics map[string]any
	if err := json.Unmarshal(body, &metrics); err != nil {
		return nil, fmt.Errorf("error unmarshaling metrics: %v", err)
	}
	return metrics, nil
}

// Equals returns whether oldMetrics is equal to newMetrics
func (m *Metrics) Equals(oldMetrics, newMetrics map[string]any) bool {
	return reflect.DeepEqual(oldMetrics, newMetrics)
}

// Update returns new metrics if metrics were changed
func (m *Metrics) Update() {
	newMetrics, err := m.Fetch()
	if err != nil {
		log.Printf("Failed to fetch metrics: %v", err)
		m.notifier.SendBroadcastMessage(fmt.Sprintf("Failed to update metrics: %v", err))
	}

	if m.lastMetrics == nil || !m.Equals(m.lastMetrics, newMetrics) {
		log.Println("Metrics have changed, sending update notification...")
		m.sendMetrics(newMetrics)
		m.lastMetrics = newMetrics
	}

	log.Println("No changes in metrics.")
}

func (m *Metrics) sendMetrics(metrics map[string]any) {
	var message string
	keys := make([]string, 0, len(metrics))
	for key := range metrics {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		message += fmt.Sprintf("%s: %s\n", key, metrics[key])
	}
	m.notifier.SendBroadcastMessage(message)
}
