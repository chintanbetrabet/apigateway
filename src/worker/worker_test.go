package worker

import (
	"testing"

	`github.com/qubole/edith/pkg/apps/edith`
	`github.com/qubole/edith/pkg/command`
	"github.com/qubole/edith/pkg/spark"
)

var (
	testRedisServer = "127.0.0.1:6379"
)

func TestWorker_pop(t *testing.T) {
	type fields struct {
	}
	type args struct {
		queue string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *edith.Command
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "TestSucces",
			fields:  fields{},
			args:    args{queue: "testqueue"},
			want:    &edith.Command{ID: 1, Type: "spark_app", SparkApp: &spark.App{ID: 1}, Info: &command.RunInfo{Operation: "create"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		w := testNewWorker(tt.args.queue)
		stopChan := make(chan struct{})

		t.Run(tt.name+"Init", func(t *testing.T) {
			t.Parallel()
			w.loopPushCmd(stopChan)
		})

		t.Run(tt.name+"Push", func(t *testing.T) {
			t.Parallel()
			w.push(tt.want)
		})

		t.Run(tt.name+"Pop", func(t *testing.T) {
			t.Parallel()

			var got command.Command
			var err error
			for {
				got, err = w.pop()
				if got != nil {
					break
				}
			}

			stopChan <- struct{}{}

			if (err != nil) != tt.wantErr {
				t.Errorf("Worker.pop() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.GetID() != tt.want.ID {
				t.Errorf("Worker.pop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func testNewWorker(queue string) *Worker {
	cmdManager := &edith.Codec{}
	return newWorker(1, testRedisServer, cmdManager, Logger("info"), CmdQueue(queue))
}
