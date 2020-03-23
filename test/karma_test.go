package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	networkutils "github.com/mesosphere/ksphere-testing-framework/pkg/utils/networking"
	"github.com/mesosphere/kubeaddons/pkg/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	karmaAlertsPath    = "/ops/portal/kommander/monitoring/karma/alerts.json"
	karmaHealthyStatus = "success"
	karmaPodPrefix     = "kommander-kubeaddons-karma-"
	karmaPort          = "8080"
	ns                 = "kommander"
)

// -----------------------------------------------------------------------------
// Karma - Tests
// -----------------------------------------------------------------------------

func karmaChecker(t *testing.T, cluster test.Cluster) test.Job {
	return func(t *testing.T) error {
		karmaPod, err := findKarmaPod(cluster)
		if err != nil {
			return fmt.Errorf("could not find karma pod in cluster %s: %s", cluster.Name(), err)
		}
		if karmaPod.Status.Phase != corev1.PodRunning {
			return fmt.Errorf("karma pod is not running, it's in phase %s", karmaPod.Status.Phase)
		}

		// create a port-forward so we can access the karma alerts api
		localport, stop, err := networkutils.PortForward(cluster.Config(), karmaPod.Namespace, karmaPod.Name, karmaPort)
		if err != nil {
			return fmt.Errorf("could not set up port forward for pod/%s port %s: %s", karmaPod.Name, karmaPort, err)
		}

		// make a request to karma's alerts api to check current alerts status
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/%s", localport, karmaAlertsPath))
		if err != nil {
			return fmt.Errorf("could not GET karma alerts api at pod/%s port %s path %s: %s", karmaPod.Name, karmaPort, karmaAlertsPath, err)
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("could not read response body from karma API: %s", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("karma alert api failed [status code: %d]: %s", resp.StatusCode, b)
		}

		alerts := map[string]interface{}{}
		if err := json.Unmarshal(b, &alerts); err != nil {
			return fmt.Errorf("could not decode JSON from karma API: %s", err)
		}

		status, ok := alerts["status"].(string)
		if !ok {
			return fmt.Errorf("could not decode karma alert status: %+v", alerts)
		}
		if status != karmaHealthyStatus {
			return fmt.Errorf("karma alerts status unhealthy [%s]: %+v", status, alerts)
		}

		// cleanup
		close(stop)
		t.Logf("INFO: successfully tested the karma alerts api 🖒")

		return nil
	}
}

// -----------------------------------------------------------------------------
// Karma - Test Utils
// -----------------------------------------------------------------------------

func findKarmaPod(cluster test.Cluster) (*corev1.Pod, error) {
	pods, err := cluster.Client().CoreV1().Pods(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	karmaPods := []corev1.Pod{}
	for _, pod := range pods.Items {
		if pod.ObjectMeta.GetDeletionTimestamp() != nil {
			continue
		}
		if strings.HasPrefix(pod.Name, karmaPodPrefix) {
			if len(karmaPods) > 0 {
				return nil, fmt.Errorf("found multiple running karma pods, expected 1 [%s, %s]", karmaPods[0].Name, pod.Name)
			}
			karmaPods = append(karmaPods, pod)
		}
	}

	return &karmaPods[0], nil
}

/* sample alert from karma:

{
  "status": "success",
  "timestamp": "2020-03-19T17:47:38.645732612Z",
  "version": "v0.50",
  "upstreams": {
    "counters": {
      "total": 1,
      "healthy": 0,
      "failed": 1
    },
    "instances": [
      {
        "name": "placeholder",
        "uri": "https://placeholder.invalid",
        "publicURI": "https://placeholder.invalid",
        "headers": {},
        "error": "Get https://placeholder.invalid/api/v2/status: dial tcp: lookup placeholder.invalid on 10.0.0.10:53: no such host",
        "version": "",
        "cluster": "12abdd31cf62e1aa0544a8ee68e9bddeef51ff5f",
        "clusterMembers": [
          "placeholder"
        ]
      }
    ],
    "clusters": {
      "12abdd31cf62e1aa0544a8ee68e9bddeef51ff5f": [
        "placeholder"
      ]
    }
  },
  "silences": {
    "12abdd31cf62e1aa0544a8ee68e9bddeef51ff5f": {}
  },
  "groups": [],
  "totalAlerts": 0,
  "colors": {},
  "filters": [],
  "counters": [],
  "settings": {
    "staticColorLabels": [
      "job"
    ],
    "annotationsDefaultHidden": false,
    "annotationsHidden": [
      "help"
    ],
    "annotationsVisible": null,
    "sorting": {
      "grid": {
        "order": "startsAt",
        "reverse": true,
        "label": "alertname"
      },
      "valueMapping": {}
    },
    "silenceForm": {
      "strip": {
        "labels": []
      },
      "author": ""
    },
    "alertAcknowledgement": {
      "enabled": false,
      "durationSeconds": 900,
      "author": "karma",
      "commentPrefix": "ACK!"
    }
  }
}

*/
