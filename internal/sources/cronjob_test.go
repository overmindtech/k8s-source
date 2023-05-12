package sources

import (
	"context"
	"testing"
	"time"
)

var cronJobYAML = `
apiVersion: batch/v1
kind: CronJob
metadata:
  name: my-cronjob
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: my-container
            image: alpine
            command: ["/bin/sh", "-c"]
            args:
            - sleep 10; echo "Hello, world!"
          restartPolicy: OnFailure
`

func TestCronJobSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := NewCronJobSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:        &source,
		GetQuery:      "my-cronjob",
		GetScope:      sd.String(),
		SetupYAML:     cronJobYAML,
		GetQueryTests: QueryTests{},
	}

	st.Execute(t)

	// Additionally, make sure that the job has a link back to the cronjob that
	// created it
	jobSource := NewJobSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	// Wait for the job to be created
	err := WaitFor(10*time.Second, func() bool {
		jobs, err := jobSource.List(context.Background(), sd.String())

		if err != nil {
			t.Fatal(err)
			return false
		}

		// Ensure that the job has a link back to the cronjob
		for _, job := range jobs {
			for _, q := range job.LinkedItemQueries {
				if q.Query == "my-cronjob" {
					return true
				}
			}

		}

		return false
	})

	if err != nil {
		t.Fatal(err)
	}
}
