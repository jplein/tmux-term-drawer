package window

import (
	"encoding/json"
	"io"
	"os"
	"path"
)

const WindowMapPrefix = ".term-drawer-pane-map"

type Map struct {
	// The process ID of the running tmux server
	Pid int `json:"pid,omitempty"`

	// A map of windows to drawer panes. Each key is the ID of a tmux window, each
	// value is the ID of the drawer pane for that window.
	Panes map[string]string `json:"panes,omitempty"`

	Socket string
}

func (m *Map) FromJSON(j []byte) error {
	return json.Unmarshal(j, m)
}

func (m *Map) configFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var filename string

	if m.Socket != "" {
		filename = WindowMapPrefix + "-" + m.Socket + ".json"
	} else {
		filename = WindowMapPrefix + ".json"
	}

	return path.Join(home, filename), nil
}

func (m *Map) fromFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	j, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	return m.FromJSON(j)
}

func (m *Map) Read() error {
	file, err := m.configFile()
	if err != nil {
		return err
	}

	return m.fromFile(file)
}

func (m *Map) toFile(file string) error {
	buf, err := json.Marshal(m)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	_, err = f.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

func (m *Map) Write() error {
	file, err := m.configFile()
	if err != nil {
		return err
	}

	return m.toFile(file)
}

func (m *Map) Clear() error {
	m.Pid = 0
	m.Panes = nil

	return m.Write()
}

// Ensure the configuration file exists. If it does not, create an empty one.
func (m *Map) Initialize() error {
	var err error

	var config string
	if config, err = m.configFile(); err != nil {
		return err
	}

	if _, err = os.Stat(config); err != nil {
		// file does not exist, create it
		if err = m.Clear(); err != nil {
			return err
		}
	}

	return nil
}

func (m *Map) GetPid() int {
	return m.Pid
}

func (m *Map) SetPid(pid int) {
	m.Pid = pid
}

func (m *Map) GetPane(window string) string {
	if m.Panes != nil {
		return m.Panes[window]
	}
	return ""
}

func (m *Map) SetPane(window, pane string) {
	if m.Panes == nil {
		m.Panes = make(map[string]string)
	}

	m.Panes[window] = pane
}
