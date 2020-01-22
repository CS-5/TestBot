package main

import (
	"fmt"
	"io/ioutil"

	"github.com/tidwall/gjson"
)

type botConfig struct {
	requestableRoles []string
	simpleCommands   map[string]string
	permissions      map[string][]string
}

func getConfig(path string) (*botConfig, error) {
	json, err := getJSON(path)
	if err != nil {
		return &botConfig{}, err
	}

	requestableRoles := getRequestableRoles(json)
	simpleCommands, err := getSimpleCommands(json)
	if err != nil {
		return &botConfig{}, err
	}

	permissions, err := getPermissions(json)
	if err != nil {
		return &botConfig{}, err
	}

	return &botConfig{
		requestableRoles: requestableRoles,
		simpleCommands:   simpleCommands,
		permissions:      permissions,
	}, nil
}

func getJSON(path string) (string, error) {
	file, err := initFile(path)
	if err != nil {
		return "", err
	}

	json, err := ioutil.ReadFile(file.Name())
	if err != nil {
		return "", err
	}

	return string(json), nil
}

func getRequestableRoles(json string) []string {
	var ids []string
	for _, id := range gjson.Get(json, "requestableRoles").Array() {
		ids = append(ids, id.String())
	}

	return ids
}

func getSimpleCommands(json string) (map[string]string, error) {
	out := make(map[string]string)

	m, ok := gjson.Parse(
		gjson.Get(json, "simpleCommands").String(),
	).Value().(map[string]interface{})
	if !ok {
		return out,
			fmt.Errorf("unable to get list of simple commands from config file")
	}

	/* TODO: Better error handling in the event of a non-string value*/
	for k, v := range m {
		out[k] = v.(string)
	}

	return out, nil
}

/* TODO: I feel like this function is really a disaster and needs a rework sometime */
func getPermissions(json string) (map[string][]string, error) {
	out := make(map[string][]string)

	r := gjson.Get(json, "permissions")
	r.ForEach(func(key, value gjson.Result) bool {
		var roles []string

		if value.IsArray() {
			for _, v := range value.Array() {
				roles = append(roles, v.String())
			}
		}

		out[key.String()] = roles

		return true
	})
	return out, nil
}
