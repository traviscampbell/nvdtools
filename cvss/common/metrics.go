// Copyright (c) Facebook, Inc. and its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

const (
	partSeparator   = "/"
	metricSeparator = ":"
)

// Metrics maps value to metrics
type Metrics map[string]string

// implement parts of Vector for Metrics

func (ms Metrics) Get(m string) (string, error) {
	if value, ok := ms[m]; ok {
		return value, nil
	}
	return "", fmt.Errorf("metric %q not defined", m)
}

func (ms Metrics) Set(metric string, value string) error {
	ms[metric] = value
	return nil
}

func (ms Metrics) String() string {
	var parts []string
	for metric, value := range ms {
		parts = append(parts, fmt.Sprintf("%s%s%s", metric, metricSeparator, value))
	}
	return strings.Join(parts, partSeparator)
}

// parse A:B/C:D into map{A:B, C:D}
func strToMetrics(str string) (Metrics, error) {
	metrics := make(Metrics)
	for _, part := range strings.Split(str, partSeparator) {
		tmp := strings.Split(part, metricSeparator)
		if len(tmp) != 2 {
			return nil, fmt.Errorf("need two values separated by %s, got %q", metricSeparator, part)
		}
		if _, exists := metrics[tmp[0]]; exists {
			return nil, fmt.Errorf("metric %q already set", tmp[0])
		}
		metrics[tmp[0]] = tmp[1]
	}
	return metrics, nil
}

// WeightsMetrics uses weights to do Set and Parse metrics
type WeightsMetrics struct {
	Metrics
	Weights map[string]map[string]float64
}

func (wms WeightsMetrics) Set(metric string, value string) error {
	values, ok := wms.Weights[metric]
	if !ok {
		return fmt.Errorf("metric %q not defined for vector", metric)
	}
	if _, ok = values[value]; !ok {
		return fmt.Errorf("can't set metric %q to %q", metric, value)
	}
	return wms.Metrics.Set(metric, value)
}

func (wms WeightsMetrics) Parse(str string) error {
	metrics, err := strToMetrics(str)
	if err != nil {
		return errors.Wrapf(err, "unable to parse metrics")
	}
	for metric, value := range metrics {
		if err = wms.Set(metric, value); err != nil {
			return errors.Wrapf(err, "unable to set metric %q to %q", metric, value)
		}
	}
	return nil
}

// weight functions

func (wms WeightsMetrics) Weight(metric string) (float64, error) {
	value, err := wms.Get(metric)
	if err != nil {
		return 0, errors.Wrapf(err, "unable to get value for metric %q", metric)
	}
	// this will always work because we only set values for valid metrics
	return wms.Weights[metric][value], nil
}

func (wms WeightsMetrics) WeightMust(metric string) float64 {
	w, err := wms.Weight(metric)
	if err != nil {
		panic(err)
	}
	return w
}

func (wms WeightsMetrics) WeightDefault(metric string, def float64) float64 {
	if w, err := wms.Weight(metric); err == nil {
		return w
	}
	return def
}
