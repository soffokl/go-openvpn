/*
 * Copyright (C) 2018 The "MysteriumNetwork/go-openvpn" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package server

type ClientEventType string

const (
	Connect     = ClientEventType("CONNECT")
	Reauth      = ClientEventType("REAUTH")
	Established = ClientEventType("ESTABLISHED")
	Disconnect  = ClientEventType("DISCONNECT")
	Address     = ClientEventType("ADDRESS")
	//pseudo event type ENV - that means some of above defined events are multiline and ENV messages are part of it
	Env = ClientEventType("ENV")
	//constant which means that id of type int is undefined
	Undefined = -1
)

type ClientEvent struct {
	EventType ClientEventType
	ClientID  int
	ClientKey int
	Env       map[string]string
}

var UndefinedEvent = ClientEvent{
	ClientID:  Undefined,
	ClientKey: Undefined,
	Env:       make(map[string]string),
}
