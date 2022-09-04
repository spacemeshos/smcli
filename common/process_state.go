package common

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/shirou/gopsutil/process"
	"github.com/spf13/cobra"
)

var stateSingleton StateProvider
var stateSingletonOnce sync.Once

func NewStateProvider() *StateProvider {
	stateSingletonOnce.Do(func() {
		stateSingleton = StateProvider{
			latestState: defaultState(),
			mu:          &sync.Mutex{},
		}
		stateSingleton.loadStateFromFileLocked()
	})
	return &stateSingleton
}

type StateProvider struct {
	mu          *sync.Mutex
	latestState state
}

// loadState loads the state from the state file.
func (sp *StateProvider) loadStateFromFileLocked() {
	// if state file doesn't exists, create a new one.
	if _, err := os.Stat(StateFile()); os.IsNotExist(err) {
		sp.latestState = defaultState()
		sp.saveStateToFileLocked()
		return
	}
	stateBytes, err := os.ReadFile(StateFile())
	cobra.CheckErr(err)
	err = json.Unmarshal(stateBytes, &sp.latestState)
	cobra.CheckErr(err)
}

// saveState saves the latestState to the state file.
func (sp *StateProvider) saveStateToFileLocked() {
	jsonState, err := json.Marshal(sp.latestState)
	cobra.CheckErr(err)
	err = os.WriteFile(StateFile(), []byte(jsonState), 0660)
	cobra.CheckErr(err)
}

// state persists the last known state of the node.
type state struct {
	Pid int `json:"pid"`
}

func defaultState() state {
	return state{
		// Process
		Pid: -1,
	}
}

func (sp *StateProvider) NodeIsRunning() bool {
	sp.mu.Lock()
	sp.mu.Unlock()
	sp.loadStateFromFileLocked()
	if sp.latestState.Pid == -1 {
		return false
	}
	_, err := process.NewProcess(int32(sp.latestState.Pid))
	if err != nil {
		switch err {
		case process.ErrorProcessNotRunning:
			sp.latestState.Pid = -1
			sp.saveStateToFileLocked()
			return false
		default:
			cobra.CheckErr(err)
		}
	}
	return true
}

func (sp *StateProvider) GetNodePid() int {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.loadStateFromFileLocked()
	return sp.latestState.Pid
}

func (sp *StateProvider) UpdateNodePid(pid int) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.loadStateFromFileLocked()
	sp.latestState.Pid = pid
	sp.saveStateToFileLocked()
	return nil
}
