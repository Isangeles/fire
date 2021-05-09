/*
 * users.go
 *
 * Copyright (C) 2020-2021 Dariusz Sikora <dev@isangeles.pl>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */

// Package for loading server data.
package data

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/isangeles/flame/data/text"

	"github.com/isangeles/fire/data/res"
	"github.com/isangeles/fire/user"
)

const (
	userConfFile = ".user"
)

var (
	users = make(map[string]*user.User)
)

// User returns user with specified ID, or
// nil if no such user were found.
func User(id string) *user.User {
	return users[id]
}

// LoadUsers loads all users from directory
// with specified path.
func LoadUsers(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("unable to read users dir: %v",
			err)
	}
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		u, err := loadUser(filepath.Join(path, f.Name()))
		if err != nil {
			log.Printf("unable to load user: %s: %v",
				f.Name(), err)
			continue
		}
		users[u.ID()] = u
	}
	return nil
}

// SaveUsers saves all loaded users under directory
// with specified path.
func SaveUsers(path string) error {
	for _, u := range users {
		userPath := filepath.Join(path, u.ID())
		err := saveUser(userPath, u)
		if err != nil {
			return fmt.Errorf("unable to save user: %v",
				err)
		}
	}
	return nil
}

// loadUser loads user from directory with
// specified path.
func loadUser(path string) (*user.User, error) {
	userFile, err := os.Open(filepath.Join(path, userConfFile))
	if err != nil {
		return nil, fmt.Errorf("unable to open user file: %v",
			err)
	}
	userConf, err := text.UnmarshalConfig(userFile)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal user config: %v",
			err)
	}
	userData := res.UserData{
		ID:    filepath.Base(path),
		Chars: make(map[string]string),
	}
	if len(userConf["pass"]) > 0 {
		userData.Pass = userConf["pass"][0]
	}
	if len(userConf["admin"]) > 0 {
		userData.Admin = userConf["admin"][0] == "true"
	}
	for _, sid := range userConf["chars"] {
		serialID := strings.Split(sid, "#")
		if len(serialID) > 1 {
			userData.Chars[serialID[0]] = serialID[1]
		}
	}
	userData.CharFlags = userConf["char-flags"]
	return user.New(userData), nil
}

// saveUser saves user under specified path.
func saveUser(path string, user *user.User) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("unable to create user directory: %v", err)
	}
	conf := make(map[string][]string)
	conf["pass"] = []string{user.Pass()}
	conf["admin"] = []string{fmt.Sprintf("%v", user.Admin())}
	for _, c := range user.Chars() {
		serialID := fmt.Sprintf("%s#%s", c.ID, c.Serial)
		conf["chars"] = append(conf["chars"], serialID)
	}
	for _, f := range user.CharFlags() {
		conf["char-flags"] = append(conf["char-flags"], string(f))
	}
	confText := text.MarshalConfig(conf)
	confPath := filepath.Join(path, userConfFile)
	confFile, err := os.Create(confPath)
	if err != nil {
		return fmt.Errorf("unable to create user config file: %v", err)
	}
	defer confFile.Close()
	writer := bufio.NewWriter(confFile)
	writer.WriteString(confText)
	writer.Flush()
	return nil
}
