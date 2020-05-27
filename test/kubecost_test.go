package test

import (
	"fmt"
	"testing"
	"time"

	testcluster "github.com/mesosphere/ksphere-testing-framework/pkg/cluster"
	testharness "github.com/mesosphere/ksphere-testing-framework/pkg/harness"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func kubecostChecker(t *testing.T, cluster testcluster.Cluster) testharness.Job {
	return func(t *testing.T) error {
		if err := kubectl("apply", "-f", "./artifacts/kubecost-checker.yaml"); err != nil {
			return err
		}

		succeeded := false
		timeout := time.Now().Add(time.Minute * 1)
		for timeout.After(time.Now()) {
			job, err := cluster.Client().BatchV1().Jobs("default").Get("kubecost-checker", metav1.GetOptions{})
			if err != nil {
				return err
			}
			if job.Status.Succeeded == 1 {
				succeeded = true
				break
			}
			time.Sleep(time.Second * 1)
		}

		if !succeeded {
			return fmt.Errorf("kubecost checker job did not succeed within timeout")
		}
		t.Log("kubecost checker job succeeded 🙃")
		return nil
	}
}
