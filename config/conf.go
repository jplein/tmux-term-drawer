package window

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type Position string

const (
	Top    Position = "top"
	Bottom Position = "bottom"
	Left   Position = "left"
	Right  Position = "right"
)

const DefaultPosition = Right

type Units string

const (
	Absolute Units = "absolute"
	Percent  Units = "percent"
)

const DefaultUnits = Percent

type Size int

const DefaultSize = 30

const DefaultSessionName = "term-drawer"

type Config struct {
	Position    Position `yaml:"position"`
	Size        Size     `yaml:"size"`
	Units       Units    `yaml:"units"`
	SessionName string   `yaml:"sessionName"`
}

type ConfigDoc struct {
	TmuxTermDrawer Config `yaml:"tmux-term-drawer"`
}

func (c *Config) Init() {
	c.Position = DefaultPosition
	c.Size = DefaultSize
	c.Units = DefaultUnits
	c.SessionName = DefaultSessionName
}

const ConfigFilename = ".term-drawer-config.yaml"

func (c *Config) configFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(home, ConfigFilename), nil
}

func (c *Config) configFileExists() (bool, error) {
	file, err := c.configFile()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(file)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	// If there's some other error, return it

	return false, err
}

func (c *Config) Validate() error {
	switch {
	case c.Position != Top && c.Position != Bottom && c.Position != Left && c.Position != Right:
		return fmt.Errorf("invalid position: '%s', valid values: %s, %s, %s, %s", c.Position, Top, Bottom, Left, Right)
	case c.Size <= 0:
		return fmt.Errorf("invalid width: %d, expected number greater than 0", c.Size)
	case c.Units != Absolute && c.Units != Percent:
		return fmt.Errorf("invalid units: '%s', valid values: %s, %s", c.Units, Absolute, Percent)
	case c.SessionName == "":
		return fmt.Errorf("invalid session name: '%s', expected non-empty string", c.SessionName)
	default:
		return nil
	}
}

func (c *Config) Write() error {
	file, err := c.configFile()
	if err != nil {
		return err
	}

	doc := ConfigDoc{}
	doc.TmuxTermDrawer = *c

	buf, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}

	_, err = f.Write(buf)
	return err
}

func (c *Config) Read() error {
	// If the configuration file does not exist, use a default configuration, and
	// write that default configuration to file
	exists, err := c.configFileExists()
	if err == nil && !exists {
		c.Init()
		return c.Write()
	}

	file, err := c.configFile()
	if err != nil {
		return err
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}

	y, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	doc := ConfigDoc{}
	err = yaml.Unmarshal(y, &doc)
	if err != nil {
		return err
	}

	if doc.TmuxTermDrawer.Validate() != nil {
		return doc.TmuxTermDrawer.Validate()
	}

	c.Position = doc.TmuxTermDrawer.Position
	c.Size = doc.TmuxTermDrawer.Size
	c.Units = doc.TmuxTermDrawer.Units
	c.SessionName = doc.TmuxTermDrawer.SessionName

	return nil
}
