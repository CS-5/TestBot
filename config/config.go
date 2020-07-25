package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/PulseDevelopmentGroup/0x626f74/multiplexer"
	"github.com/PulseDevelopmentGroup/0x626f74/util"

	"github.com/tidwall/gjson"
)

type (
	// BotConfig defines the configuration container for the bot
	BotConfig struct {
		Path string

		ErrorChannel string

		SimpleCommands map[string]string
		Permissions    map[string]*multiplexer.CommandPermissions
	}

	// BotPermissions contains the permission maps for roles, channels, and
	// users based on the config file.
	BotPermissions struct {
		RoleIDs map[string][]string
		ChanIDs map[string][]string
		UserIDs map[string][]string
	}
)

// Get loads the config from the json file at the path specified
func Get(path string) (*BotConfig, error) {
	json, err := getJSON(path)
	if err != nil {
		return &BotConfig{}, err
	}

	simpleCommands, err := getSimpleCommands(json)
	if err != nil {
		return &BotConfig{}, err
	}

	perms := getPermissions(json)

	return &BotConfig{
		Path:           path,
		ErrorChannel:   gjson.Get(json, "errorChannel").String(),
		SimpleCommands: simpleCommands,
		Permissions:    perms,
	}, nil
}

func (c *BotConfig) Update() error {
	new, err := Get(c.Path)
	if err != nil {
		return err
	}

	c.Path = new.Path
	c.SimpleCommands = new.SimpleCommands
	c.Permissions = new.Permissions

	return nil
}

func getJSON(path string) (string, error) {
	var json []byte

	if !util.IsURL(path) {
		file, err := util.InitFile(path)
		if err != nil {
			return "", err
		}

		json, err = ioutil.ReadFile(file.Name())
		if err != nil {
			return "", err
		}
	} else {
		resp, err := http.Get(path)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		json, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
	}

	return string(json), nil
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

	/* TODO: Better error handling in the event of a non-string value */
	for k, v := range m {
		out[k] = v.(string)
	}

	return out, nil
}

// TODO: Implement support for getting user ids and channel ids
func getPermissions(json string) map[string]*multiplexer.CommandPermissions {
	out := make(map[string]*multiplexer.CommandPermissions)

	p := gjson.Get(json, "permissions")
	p.ForEach(func(key, value gjson.Result) bool {
		var roles []string

		if value.IsArray() {
			for _, v := range value.Array() {
				roles = append(roles, v.String())
			}
		}

		out[strings.ToLower(key.String())] = &multiplexer.CommandPermissions{
			RoleIDs: roles,
		}

		return true
	})
	return out
}
