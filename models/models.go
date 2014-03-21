// Package models includes data models for the application.
package models

import (
	"encoding/json"
	"errors"

	uuid "code.google.com/p/go-uuid/uuid"
	stor "github.com/creatorrr/toyserver/stor"
)

const (
	_WAITING = iota
	_PLAYING
	_ENDED

	allowedMembers = 6
)

type (
	User struct {
		Name string `json:"name"`
		Id   string `json:"id"`
	}

	SessionData struct {
		AppData map[string]interface{} `json:"appData"`
		Members []User                 `json:"members"`
		State   int                    `json:"state"`
	}

	Session struct {
		*stor.Model
	}
)

// SessionData implements Jsoner interface
func (d *SessionData) Json() (v []byte, err error) {
	v, err = json.Marshal(d)
	return
}

func (d *SessionData) SetJson(s []byte) (e error) {
	e = json.Unmarshal(s, &d)
	return
}

// Constructors
func NewUser(name string) *User {
	return &User{
		name,
		uuid.NewUUID().String(),
	}
}

func NewSession(key string) *Session {
	return &Session{
		&stor.Model{
			key,
			&SessionData{
				make(map[string]interface{}),
				make([]User, 0),
				_WAITING,
			},
			"session",
		},
	}
}

// Setters & Getters
func (s *Session) GetData() (d *SessionData) {
	d, _ = s.Data.(*SessionData)
	return
}

func (s *Session) SetData(d *SessionData) {
	s.Data = stor.Jsoner(d)
}

// Set and get state.
func (s *Session) State() int {
	return s.GetData().State
}

func (s *Session) SetState(state int) (e error) {
	d := s.GetData()
	d.State = state

	s.SetData(d)
	return
}

// Set and get appData.
func (s *Session) AppData() map[string]interface{} {
	return s.GetData().AppData
}

func (s *Session) SetAppData(data map[string]interface{}) (e error) {
	d := s.GetData()
	d.AppData = data

	s.SetData(d)
	return
}

// Set and get members.
func (s *Session) Members() []User {
	return s.GetData().Members
}

func (s *Session) AddMember(u *User) (e error) {
	// Check if action is valid.
	switch {
	case s.State() != _WAITING:
		e = errors.New("invalid state")
		return
	case len(s.Members()) >= allowedMembers:
		e = errors.New("over capacity")
		return
	}

	// Add member.
	d := s.GetData()
	d.Members = append(d.Members, *u)

	s.SetData(d)
	return nil
}
