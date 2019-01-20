package floodsub

import (
	"errors"
	"time"
)

var (
	HeartbeatInitialDelay = 100 * time.Millisecond
	HeartbeatInterval     = 1 * time.Second
	SubFanoutTTL          = 60 * time.Second
)

// FillDefaults fills fields that are empty.
func (c *Config) FillDefaults() {
	if c == nil {
		return
	}
	if c.GetDegree() == 0 {
		c.Degree = 6
	}
	if c.GetDegreeLow() == 0 {
		c.DegreeLow = 4
	}
	if c.GetDegreeHigh() == 0 {
		c.DegreeHigh = 12
	}
	if c.GetHistoryLen() == 0 {
		c.HistoryLen = 5
	}
	if c.GetHistoryGossip() == 0 {
		c.HistoryGossip = 3
	}
	// TODO: evaluate these "safety" hard caps
	if c.GetHistoryLen() > 10000 {
		c.HistoryLen = 10000
	}
	if c.GetHistoryGossip() > 50 {
		c.HistoryGossip = 50
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c == nil {
		return nil
	}
	c.FillDefaults()
	if c.GetDegreeLow() >= c.GetDegree() {
		return errors.New("degree_low must be less than degree")
	}
	if c.GetDegreeHigh() <= c.GetDegree() {
		return errors.New("degree_high must be greater than degree")
	}
	return nil
}
