package drawer

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/jplein/tmux"
	config "github.com/jplein/tmux-term-drawer/config"
	window "github.com/jplein/tmux-term-drawer/window"
)

func getActivePID(r *tmux.Runner) (int, error) {
	result, err := r.Run("display-message -p -F '#{pid}'")
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(result)
	if err != nil {
		return 0, err
	}

	return pid, nil
}

func makeMapFile(r *tmux.Runner) error {
	pid, err := getActivePID(r)
	if err != nil {
		return err
	}

	var m window.Map
	m.Socket = r.Config.Socket

	err = m.Read()
	if err != nil {
		err = m.Clear()
		if err != nil {
			return err
		}
	}

	if m.Pid != pid {
		if err = m.Clear(); err != nil {
			return err
		}
	}

	m.SetPid(pid)

	return m.Write()
}

func getPositionArgument(c config.Config) (string, error) {
	var positionArgument string = ""
	switch c.Position {
	case config.Top:
		positionArgument = "-b"
	case config.Left:
		positionArgument = "-b"
	case config.Bottom:
		positionArgument = ""
	case config.Right:
		positionArgument = ""
	default:
		return "", fmt.Errorf("invalid value for config.Position: %s", c.Position)
	}
	return positionArgument, nil
}

func getDrawerSize(windowWidth, windowHeight int, c config.Config) (int, error) {
	size := 0
	switch {
	case c.Units == config.Character:
		size = int(c.Size)
	case c.Units == config.Percent && (c.Position == config.Bottom || c.Position == config.Top):
		size = int(math.Round(float64(windowHeight) * (float64(c.Size) / 100.0)))
	case c.Units == config.Percent && (c.Position == config.Left || c.Position == config.Right):
		size = int(math.Round(float64(windowWidth) * (float64(c.Size) / 100.0)))
	default:
		return 0, fmt.Errorf("this should not happen: how did I get to the default case?")
	}
	return size, nil
}

func getSplitParam(c config.Config) (string, error) {
	var splitParam string
	switch c.Position {
	case config.Top:
		splitParam = "-v"
	case config.Bottom:
		splitParam = "-v"
	case config.Left:
		splitParam = "-h"
	case config.Right:
		splitParam = "-h"
	default:
		return "", fmt.Errorf("this should not happen: how did I get to the default case?")
	}
	return splitParam, nil
}

func createDrawer(r *tmux.Runner, activeWindow string) (string, error) {
	var err error

	config := config.Config{}
	err = config.Read()
	if err != nil {
		return "", err
	}

	var width, height int
	if width, height, err = r.GetWindowDimensions(activeWindow); err != nil {
		return "", err
	}

	positionArgument, err := getPositionArgument(config)
	if err != nil {
		return "", err
	}

	newDimensions, err := getDrawerSize(width, height, config)
	if err != nil {
		return "", err
	}

	splitParam, err := getSplitParam(config)
	if err != nil {
		return "", err
	}

	activePane, err := r.GetActivePane()
	if err != nil {
		return "", err
	}

	workingDirectory, err := getPaneWorkingDirectory(r, activePane)
	if err != nil {
		return "", err
	}

	var output string
	cmd := fmt.Sprintf(
		"split-window %s -f %s -l %d -c '%s' -P -F '#{pane_id}'",
		splitParam,
		positionArgument,
		newDimensions,
		workingDirectory,
	)

	fmt.Printf("debug: cmd: %s\n", cmd)

	if output, err = r.Run(cmd); err != nil {
		return "", err
	}

	pane := tmux.Trim(output)

	return pane, nil
}

func getSourceStringForMovePane(r *tmux.Runner, pane string) (string, error) {
	var err error

	var source string
	var cmd string = fmt.Sprintf(
		"list-panes -a -F '#{session_name}:#{window_index}.#{pane_index}' -f '#{m:#{pane_id},%s}'",
		pane,
	)
	if source, err = r.Run(cmd); err != nil {
		return "", err
	}

	return tmux.Trim(source), nil
}

func hideDrawer(r *tmux.Runner, pane string, stashSession string) error {
	var err error
	var source string

	if source, err = getSourceStringForMovePane(r, pane); err != nil {
		return err
	}
	var target = fmt.Sprintf("%s:{end}", stashSession)
	cmd := fmt.Sprintf(
		"break-pane -s '%s' -a -t '%s'",
		source,
		target,
	)

	if _, err = r.Run(cmd); err != nil {
		return err
	}
	return nil
}

func showDrawer(r *tmux.Runner, pane, activeSession, activeWindow string) error {
	var err error

	var source string
	if source, err = getSourceStringForMovePane(r, pane); err != nil {
		return err
	}

	config := config.Config{}
	err = config.Read()
	if err != nil {
		return err
	}

	var width, height int
	if width, height, err = r.GetWindowDimensions(activeWindow); err != nil {
		return err
	}

	positionArgument, err := getPositionArgument(config)
	if err != nil {
		return err
	}

	newDimensions, err := getDrawerSize(width, height, config)
	if err != nil {
		return err
	}

	splitParam, err := getSplitParam(config)
	if err != nil {
		return err
	}

	var dest = fmt.Sprintf("%s:%s", activeSession, activeWindow)
	var cmd string = fmt.Sprintf(
		"move-pane %s -l %d -f %s -s '%s' -t '%s'",
		splitParam,
		newDimensions,
		positionArgument,
		source,
		dest,
	)
	_, err = r.Run(cmd)
	return err
}

func getPaneExists(r *tmux.Runner, pane string) (bool, error) {
	var err error

	if pane == "" {
		return false, nil
	}

	var output string
	if output, err = r.Run("list-panes -a -F '#{pane_id}'"); err != nil {
		return false, err
	}

	exists := false

	panes := strings.Split(output, "\n")
	for _, p := range panes {
		if p == pane {
			exists = true
			break
		}
	}

	return exists, nil
}

func getPaneSession(r *tmux.Runner, pane string) (string, error) {
	// 'list-panes', '-a', '-F', '#{session_name}', '-f', `#{m:#{pane_id},${pane}}`
	var err error

	var output string
	if output, err = r.Run(fmt.Sprintf("list-panes -a -F '#{session_name}' -f '#{m:#{pane_id},%s}'", pane)); err != nil {
		return "", err
	}

	return tmux.Trim(output), nil
}

func getPaneWindow(r *tmux.Runner, pane string) (string, error) {
	// 'list-panes', '-a', '-F', '#{window_id}', '-f', `#{m:#{pane_id},${pane}}`
	var err error

	var output string
	if output, err = r.Run(fmt.Sprintf("list-panes -a -F '#{window_id}' -f '#{m:#{pane_id},%s}'", pane)); err != nil {
		return "", err
	}

	return tmux.Trim(output), nil
}

func getPaneWorkingDirectory(r *tmux.Runner, pane string) (string, error) {
	// 'list-panes', '-a', '-F', '#{pane_current_path}', '-f', `#{m:#{pane_id},${pane}}`
	var err error

	var output string
	if output, err = r.Run(fmt.Sprintf("list-panes -a -F '#{pane_current_path}' -f '#{m:#{pane_id},%s}'", pane)); err != nil {
		return "", err
	}

	return tmux.Trim(output), nil
}

func Toggle(socketName string) error {
	config := config.Config{}
	err := config.Read()
	if err != nil {
		return err
	}

	var activeSession string
	if activeSession, err = tmux.GetActiveSession(tmux.Config{Socket: socketName}); err != nil {
		return err
	}

	var r *tmux.Runner = &tmux.Runner{}
	if err = r.Init(tmux.Config{Socket: socketName}); err != nil {
		return err
	}

	defer func() {
		if err = r.Close(); err != nil {
			os.Stderr.Write([]byte(fmt.Sprintf("Error closing tmux runner: %s", err.Error())))
		}
	}()

	if err = makeMapFile(r); err != nil {
		return err
	}

	if err = r.AttachSession(activeSession); err != nil {
		return err
	}
	if err = r.StartSession(config.SessionName); err != nil {
		return err
	}

	var activeWindow string
	if activeWindow, err = r.GetActiveWindow(); err != nil {
		return err
	}

	var m window.Map
	m.Socket = socketName

	if err = m.Initialize(); err != nil {
		return err
	}
	if err = m.Read(); err != nil {
		return err
	}

	pane := m.GetPane(activeWindow)

	var lastPaneExists bool
	if lastPaneExists, err = getPaneExists(r, pane); err != nil {
		return err
	}

	var columnsBefore []tmux.Column
	if columnsBefore, err = r.ListColumns(); err != nil {
		return err
	}

	var paneVisible bool = false
	var paneSession, paneWindow string
	if lastPaneExists {
		if paneSession, err = getPaneSession(r, pane); err != nil {
			return err
		}
		if paneWindow, err = getPaneWindow(r, pane); err != nil {
			return err
		}
		if paneSession == activeSession && paneWindow == activeWindow {
			paneVisible = true
		}
	}

	if !lastPaneExists {
		// there's not a drawer pane for this window in the configuration file, or
		// there is but that pane doesn't exist, so create a new one
		if pane, err = createDrawer(r, activeWindow); err != nil {
			return err
		}
	} else if paneVisible {
		if err = hideDrawer(r, pane, config.SessionName); err != nil {
			return err
		}
	} else {
		if err = showDrawer(r, pane, activeSession, activeWindow); err != nil {
			return err
		}
	}

	var columnsAfter []tmux.Column
	if columnsAfter, err = r.ListColumns(); err != nil {
		return err
	}

	var totalBefore int = 0
	for _, col := range columnsBefore {
		if col.Pane != pane {
			totalBefore += col.Width
		}
	}

	var totalAfter int = 0
	for _, col := range columnsAfter {
		if col.Pane != pane {
			totalAfter += col.Width
		}
	}

	for _, col := range columnsBefore {
		if col.Pane != pane {
			var ratioBefore float64 = float64(col.Width) / float64(totalBefore)
			var newWidth int = int(
				math.Round(
					ratioBefore * float64(totalAfter),
				),
			)
			if err = r.SetPaneWidth(col.Pane, newWidth); err != nil {
				return err
			}
		}
	}

	m.SetPane(activeWindow, pane)
	if err = m.Write(); err != nil {
		return err
	}

	// var drawerPane string
	// if drawerPane, err = getPaneForWindow(r, activeWindow); err != nil {
	// 	return err
	// }

	return nil
}
