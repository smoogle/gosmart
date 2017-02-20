// This file is part of gosmart, a set of libraries to communicate with
// the Samsumg SmartThings API using Go (golang).
//
// http://github.com/marcopaganini/gosmart
// (C) 2016 by Marco Paganini <paganini@paganini.net>

package gosmart

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	tokenFilePrefix = ".smartthings.token"
)

// Global configuration for smart things.
type Config struct {
	ClientID, Secret string
}

// Represents all smart things.
type SmartThings struct {
	client *http.Client
	endpoint string
	Devices []Device
}

func Connect(ctx context.Context, cfg Config) (SmartThings, error) {
	st := SmartThings{}
	tokenFile := fmt.Sprintf("%s_%s.json", tokenFilePrefix, cfg.ClientID)
	config := NewOAuthConfig(cfg.ClientID, cfg.Secret)
	token, err := GetToken(tokenFile, config)
	if err != nil {
		return st, err
	}
	st.client = config.Client(ctx, token)
	st.endpoint, err = GetEndPointsURI(st.client)
	if err != nil {
		return st, err
	}
	return st, st.Refresh()
}

// Refresh all the devices that are available.
func (st *SmartThings) Refresh() error {
	all, err := GetDevices(st.client, st.endpoint)
	if err != nil {
		return err
	}
	var raw []*DeviceInfo
	for _, d := range all {
		detail, err := GetDeviceInfo(st.client, st.endpoint, d.ID)
		if err != nil {
			return err
		}
		raw = append(raw, detail)
	}
	st.Devices = nil
	for _, rd := range raw {
		nd := Device{
			st: st,
			ID: rd.ID,
			Name: rd.Name,
			DisplayName: rd.DisplayName,
			Attributes: make(map[string]float64),
		}
		for k, v := range rd.Attributes {
			fv, ok := v.(float64)
			if !ok {
				fv = 0
			}
			nd.Attributes[k] = fv
		}
		err := nd.Refresh()
		if err != nil {
			return err
		}
		st.Devices = append(st.Devices, nd)
	}
	return nil
}

// Device is a representation of a Device
type Device struct {
	st *SmartThings
	ID, Name, DisplayName string
	Attributes map[string]float64
	Commands []string
}

// Refresh the available device commands.
func (d *Device) Refresh() error {
	dcs, err := GetDeviceCommands(d.st.client, d.st.endpoint, d.ID)
	if err != nil {
		return err
	}
	cmds := make(map[string]bool)
	d.Commands = nil
	for _, dc := range dcs {
		if cmds[dc.Command] {
			continue
		}
		d.Commands = append(d.Commands, dc.Command)
		cmds[dc.Command] = true
	}
	return nil
}

func (d *Device) Call(cmd string, args ...float64) error {
	found := false
	for _, c := range d.Commands {
		if cmd == c {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("unavailable command: %v", cmd)
	}
	if len(args) > 1 {
		return errors.New("too many arguments")
	}
	path := fmt.Sprintf("/devices/%s/%s", d.ID, cmd)
	if len(args) > 0 {
		var sargs []string
		for _, a := range args {
			sargs = append(sargs, fmt.Sprintf("%v", a))
		}
		path = fmt.Sprintf("%s/%v", path, strings.Join(sargs, "/"))
	}
	_, err := issueCommand(d.st.client, d.st.endpoint, path)
	return err
}

// DeviceList holds the list of devices returned by /devices
type DeviceList struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// DeviceInfo holds information about a specific device.
type DeviceInfo struct {
	DeviceList
	Attributes map[string]interface{} `json:"attributes"`
}

// DeviceCommand holds one command a device can accept.
type DeviceCommand struct {
	Command string                 `json:"command"`
	Params  map[string]interface{} `json:"params"`
}

// GetDevices returns the list of devices from smartthings using
// the specified http.client and endpoint URI.
func GetDevices(client *http.Client, endpoint string) ([]DeviceList, error) {
	ret := []DeviceList{}

	contents, err := issueCommand(client, endpoint, "/devices")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(contents, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// GetDeviceInfo returns device specific information about a particular device.
func GetDeviceInfo(client *http.Client, endpoint string, id string) (*DeviceInfo, error) {
	ret := &DeviceInfo{}

	contents, err := issueCommand(client, endpoint, "/devices/"+id)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(contents, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// GetDeviceCommands returns a slice of commands a specific device accepts.
func GetDeviceCommands(client *http.Client, endpoint string, id string) ([]DeviceCommand, error) {
	ret := []DeviceCommand{}

	contents, err := issueCommand(client, endpoint, "/devices/"+id+"/commands")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(contents, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// issueCommand sends a given command to an URI and returns the contents
func issueCommand(client *http.Client, endpoint string, cmd string) ([]byte, error) {
	uri := endpoint + cmd
	resp, err := client.Get(uri)
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return contents, nil
}
