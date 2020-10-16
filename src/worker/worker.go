package worker

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/qubole/edith/internal/httpclient"

	"github.com/go-redis/redis"

	"github.com/qubole/edith/pkg/command"
	"github.com/qubole/edith/pkg/state"
)

// Worker struct..
type Worker struct {
	id              int
	server          string
	cmdQueue        string
	cmdQueueBackup  string
	cmdQueueTimeOut int64
	redis           *redis.Client
	logger          log.Logger
	cmdChan         chan command.Command
	cmdManager      command.Codec
	stop            <-chan struct{}
	quit            bool
}

var (
	// ErrEmptyQueue redis command queue is empty/
	ErrEmptyQueue = errors.New("empty redis queue")

	defaultLogLevel        = "info"
	defaultCmdQueueTimeout = int64(5)
	defaultCmdQueue        = "cmdqueue"
)

// New is worker constructor.
func newWorker(id int, server string, cmdManager command.Codec, options ...Option) *Worker {
	w := &Worker{
		id:         id,
		server:     server,
		cmdManager: cmdManager,
		redis: redis.NewClient(&redis.Options{
			Addr: server,
		}),
	}

	// set defaults
	opts := append([]Option{}, Logger(defaultLogLevel), CmdQueueTimeout(defaultCmdQueueTimeout), CmdQueue(defaultCmdQueue))

	// set overrides
	opts = append(opts, options...)

	// apply
	for _, opt := range opts {
		opt(w)
	}

	w.cmdChan = make(chan command.Command, 100)

	return w
}

// push cmd in queue.
func (w *Worker) push(cmd command.Command) {
	w.cmdChan <- cmd
}

// loopProcessCmd Workers to fetch commands.
func (w *Worker) loopProcessCmd(stop <-chan struct{}) error {
	for {
		select {
		case <-stop:
			_ = w.logger.Log("method", "loopProcessCmd", "status", "exit", "redis_server", w.server)
			return fmt.Errorf("worker.loopProcessCmd.interrupted")
		default:
		}

		cmd, err := w.pop()

		if err == ErrEmptyQueue {
			continue
		}

		if err != nil {
			_ = w.logger.Log("method", "loopProcessCmd", "context", "pop", "error", err)
			continue
		}

		err = w.run(cmd)
		if err != nil {
			_ = w.logger.Log("method", "loopProcessCmd", "context", "run", "error", err, "cmd", cmd.String())
			continue
		}
		_ = w.logger.Log("method", "loopProcessCmd", "context", "run", "cmd", cmd.String())
	}
}

func (w *Worker) loopPushCmd(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			_ = w.logger.Log("method", "loopPushCmd", "status", "exit", "redis_server", w.server)
			return
		case cmd := <-w.cmdChan:
			b, err := w.cmdManager.MarshalCommand(cmd)
			if err != nil {
				_ = w.logger.Log("method", "loopPushCmd", "context", "json.marshal", "error", err, "cmd", cmd.String())
				continue
			}

			_, err = w.redis.RPush(w.cmdQueue, b).Result()
			if err != nil {
				_ = w.logger.Log("method", "loopPushCmd", "context", "redis.rpush", "error", err, "cmd", cmd.String())
				continue
			}
		}
	}
}

func (w *Worker) run(cmd command.Command) error {
	t, err := cmd.GetType()
	if err != nil {
		return err
	}

	t.SetCommandID(fmt.Sprintf("%d", cmd.GetID()))
	t.SetLogger(w.logger)

	op := strings.ToLower(cmd.RunInfo().Operation)
	var status *state.Status
	switch op {
	case "run", "create":
		status, err = t.Create()
	case "status", "get":
		status, err = t.Get(cmd.RunInfo().GetStartOperation())
	case "cancel", "delete":
		status, err = t.Delete()
	default:
		return fmt.Errorf("worker.run.invalid_command: %s", op)
	}
	if err == nil {
		w.processStatus(cmd, status)
	} else {
		w.processError(cmd, err)
	}

	return err
}

// processStatus of command.
func (w *Worker) processStatus(cmd command.Command, status *state.Status) {
	cmdId, cmdRunInfo := cmd.GetID(), cmd.RunInfo()
	var err error
	_ = w.logger.Log("cmdID", cmdId, "method", "processStatus", "context", "begin", "state", status.State, "startOp", cmdRunInfo.GetStartOperation())
	switch status.State {
	case state.Running:
		if cmdRunInfo.CurrentState != state.Running {
			if err = w.storeState(cmd, status, cmdRunInfo.RunningHook); err == nil {
				cmdRunInfo.CurrentState = state.Running
			} else {
				_ = w.logger.Log("cmdID", cmdId, "method", "processStatus", "context", "storeState", "callback", cmdRunInfo.RunningHook, "error", err.Error())
			}
		}

	case state.Pending:
		// TODO: remove if block
		if cmdRunInfo.CurrentState != state.Pending {
			cmdRunInfo.PendingHook = cmdRunInfo.Callbacks["started"]
			status.Payload = state.Payload{Time: time.Now().String(), Message: state.Message{Pid: -1}}
			_ = w.logger.Log("cmdID", cmdId, "method", "processStatus", "context", "storeState", "callback", cmdRunInfo.PendingHook, "cmd_new_state", status.State)
			if err = w.storeState(cmd, status, cmdRunInfo.PendingHook); err == nil {
				cmdRunInfo.CurrentState = state.Pending
			}
		}
	case state.Errored, state.Terminated:
		if status.State == state.Errored {
			cmdRunInfo.ExitCode = 1
		}
		cmdRunInfo.FinishHook = cmdRunInfo.Callbacks["finished"]
		status.Payload = state.Payload{Time: time.Now().String(), Message: state.Message{Pid: -1, WrapperExitCode: cmdRunInfo.ExitCode}}
		cmdRunInfo.CurrentState = status.State
		if err = w.storeState(cmd, status, cmdRunInfo.FinishHook); err == nil {
			return
		}
		_ = w.logger.Log("cmdID", cmdId, "method", "processStatus", "context", "storeState", "callback", cmdRunInfo.FinishHook, "error", err.Error())
	}

	if cmdRunInfo.IsFinished(status.State) {
		cmdRunInfo.FinishHook = cmdRunInfo.Callbacks["finished"]
		status.Payload = state.Payload{Time: time.Now().String(), Message: state.Message{Pid: -1, WrapperExitCode: cmdRunInfo.ExitCode}}
		if err = w.storeState(cmd, status, cmdRunInfo.FinishHook); err == nil {
			return
		}
		_ = w.logger.Log("cmdID", cmdId, "method", "processStatus", "context", "storeState", "callback", cmdRunInfo.FinishHook, "error", err.Error())
		return
	}

	// log for bookkeeping purpose.
	if err != nil {
		_ = w.logger.Log("cmdID", cmdId, "method", "processStatus", "context", "storeState", "error", err, "cmd_new_state", status.State)
	}

	cmdRunInfo.Operation = "status"
	<-time.NewTicker(expBackoff(1, 2*time.Second, 7*time.Second)).C
	w.push(cmd)
}

// processError of command.
// TODO add redundancy if things fail here.
// Need to backup and rerun
func (w *Worker) processError(cmd command.Command, err error) {
	if cmd.RunInfo().RetryCount < cmd.RunInfo().MaxRetries {
		cmd.RunInfo().RetryCount++

		// randomize time to enqueue.
		<-time.NewTicker(expBackoff(1, 2*time.Second, 7*time.Second)).C
		w.push(cmd)

		return
	}
	cmd.RunInfo().FinishHook = cmd.RunInfo().Callbacks["finished"]
	cmd.RunInfo().ExitCode = 1
	payload := state.Payload{Time: time.Now().String(), Message: state.Message{Pid: -1, WrapperExitCode: cmd.RunInfo().ExitCode}}

	if err := w.storeState(cmd, &state.Status{CommandID: cmd.GetID(), State: state.Errored, Payload: payload}, cmd.RunInfo().FinishHook); err != nil {
		_ = w.logger.Log("cmdID", cmd.GetID(), "method", "processError", "context", "storeState", "error", err, "cmd", cmd.String())
	}
}

// storeState in redis/external-store.
func (w *Worker) storeState(cmd command.Command, status *state.Status, hook string) error {
	if hook == "" {
		_ = w.logger.Log("cmdID", cmd.GetID(), "method", "storeState", "context", "", "error", "blank.hook")
		return nil
	}
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Accept"] = "*/*"

	// fmt.Printf("Sending payload: %+v\n", status.Payload)

	st, _, err := httpclient.Post(context.Background(), hook, status.Payload, headers)
	_ = w.logger.Log("cmdID", cmd.GetID(), "method", "storeState", "context", "post", "hook", hook, "http.status", st, "cmd", cmd.String())
	if err != nil {
		_ = w.logger.Log("cmdID", cmd.GetID(), "method", "storeState", "context", "post", "error", err, "http.status", st, "cmd", cmd.String())
	}

	return err
}

func (w *Worker) pop() (command.Command, error) {
	var result []string
	var err error

	result, err = w.redis.BLPop(time.Duration(w.cmdQueueTimeOut)*time.Second, w.cmdQueue).Result()

	if err == redis.Nil {
		return nil, ErrEmptyQueue
	}
	if err != nil {
		return nil, err
	}

	if v, err := w.cmdManager.UnMarshalCommand([]byte(result[1])); err != nil {
		return nil, err
	} else {
		return v, nil
	}
}

func expBackoff(retry int, minBackoff, maxBackoff time.Duration) time.Duration {
	if retry < 0 {
		retry = 0
	}

	backoff := minBackoff << uint(retry)
	if backoff > maxBackoff || backoff < minBackoff {
		backoff = maxBackoff
	}

	if backoff == 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(backoff)))
}

// rmBackup removes command from backup processing queue.
// TODO - incomplete
// func (w *Worker) rmBackup(cmd *command.EdithCommand) error {
// 	// randomize time to enqueue.
// 	<-time.NewTicker(expBackoff(1, 2*time.Second, 7*time.Second)).C

// 	if _, err := w.redis.LRem(w.cmdQueueBackup, 1, w.cmdIDPrefix(cmd)).Result(); err != nil {
// 		return err
// 	}

// 	return w.rmBackup(cmd)
// }

// func (w *Worker) cmdIDPrefix(cmd *command.EdithCommand) string {
// 	return fmt.Sprintf("%d:", cmd.ID)
// }

// func (w *Worker) stripCmdIDPrefix(result string) []byte {
// 	return stripRegex.ReplaceAll([]byte(result), []byte(""))
// }
