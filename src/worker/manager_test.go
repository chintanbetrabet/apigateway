package worker

import (
	`fmt`
	"testing"

	`github.com/qubole/edith/pkg/apps/edith`
	`github.com/qubole/edith/pkg/command`
	"github.com/qubole/edith/pkg/spark"
)

func TestManager_Push(t *testing.T) {
	t.Parallel()
	type fields struct {
		m *Manager
	}
	type args struct {
		cmd *edith.Command
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *edith.Command
		wantErr bool
	}{
		{
			name:    "TestSuccess",
			fields:  fields{m: testCreateManager(2)},
			args:    args{cmd: &edith.Command{ID: 2344, Type: "spark_app", SparkApp: &spark.App{ID: 1}, Info: &command.RunInfo{Operation: "get"}}},
			want:    &edith.Command{ID: 2344, Type: "spark_app", SparkApp: &spark.App{ID: 1}, Info: &command.RunInfo{Operation: "get"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		stopChan := make(chan struct{})
		exit := false

		for _, fn := range tt.fields.m.Runnables() {
			t.Run("Runnable ", getRunnableMethod(fn, &exit, stopChan))
		}


		t.Run(tt.name + " PushCommand", func(t *testing.T) {
			t.Parallel()
			tt.fields.m.Push(tt.args.cmd)
		})

		t.Run(tt.name + " ValidateWorker", func(t *testing.T) {
			t.Parallel()
			var got command.Command
			var err error
			for {
				idx := tt.args.cmd.ID % uint64(len(tt.fields.m.workers))
				got, err = tt.fields.m.workers[idx].pop()
				if got != nil {
					break
				} else {
					fmt.Println("Got Nil")
				}
			}

			exit = true
			close(stopChan)

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

func testCreateManager(n int) *Manager {
	var ss []string
	for i := 0; i < n; i++ {
		ss = append(ss, testRedisServer)
	}
	cmdManager := &edith.Codec{}
	return NewManager(ss, cmdManager, Logger("nil"))
}

func getRunnableMethod(fn RunFn, exit *bool, stopChan <-chan struct{}) func (t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		if err := fn(stopChan); err != nil && *exit == false {
			t.Errorf("Error in executing runnable: %s", err.Error())
		}
	}
}
