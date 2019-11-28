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

package filter

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/mysteriumnetwork/go-openvpn/openvpn/log"
	"github.com/mysteriumnetwork/go-openvpn/openvpn/management"
)

const filterLANTemplate = `client-pf {{.ClientID}}
[CLIENTS DROP]
[SUBNETS ACCEPT]
{{- range $subnet := .Subnets}}
-{{$subnet}}
{{- end}}
[END]
END
`

var filterLAN = template.Must(template.New("filter_lan").Parse(filterLANTemplate))

type middleware struct {
	// TODO: consider implementing event channel to communicate required callbacks
	commandWriter management.CommandWriter
	currentEvent  clientEvent
	filter        []string
}

// NewMiddleware creates server user_auth challenge authentication middleware
func NewMiddleware(filter ...string) *middleware {
	return &middleware{
		commandWriter: nil,
		currentEvent:  undefinedEvent,
		filter:        filter,
	}
}

type clientEventType string

const (
	connect     = clientEventType("CONNECT")
	reauth      = clientEventType("REAUTH")
	established = clientEventType("ESTABLISHED")
	disconnect  = clientEventType("DISCONNECT")
	address     = clientEventType("ADDRESS")
	//pseudo event type ENV - that means some of above defined events are multiline and ENV messages are part of it
	env = clientEventType("ENV")
	//constant which means that id of type int is undefined
	undefined = -1
)

type clientEvent struct {
	eventType clientEventType
	clientID  int
	clientKey int
	env       map[string]string
}

var undefinedEvent = clientEvent{
	clientID:  undefined,
	clientKey: undefined,
	env:       make(map[string]string),
}

func (m *middleware) Start(commandWriter management.CommandWriter) error {
	m.commandWriter = commandWriter
	return nil
}

func (m *middleware) Stop(commandWriter management.CommandWriter) error {
	return nil
}

func (m *middleware) ConsumeLine(line string) (bool, error) {
	if !strings.HasPrefix(line, ">CLIENT:") {
		return false, nil
	}

	clientLine := strings.TrimPrefix(line, ">CLIENT:")

	eventType, eventData, err := parseClientEvent(clientLine)
	if err != nil {
		return true, err
	}

	switch eventType {
	case connect, reauth:
		ID, key, err := parseIDAndKey(eventData)
		if err != nil {
			return true, err
		}

		m.startOfEvent(eventType, ID, key)
	case env:
		if strings.ToLower(eventData) == "end" {
			m.endOfEvent()
			return true, nil
		}
	}

	return true, nil
}

func (m *middleware) startOfEvent(eventType clientEventType, clientID int, keyID int) {
	m.currentEvent.eventType = eventType
	m.currentEvent.clientID = clientID
	m.currentEvent.clientKey = keyID
}

func (m *middleware) endOfEvent() {
	m.handleClientEvent(m.currentEvent)
	m.reset()
}

func (m *middleware) reset() {
	m.currentEvent = undefinedEvent
}

func (m *middleware) handleClientEvent(event clientEvent) {
	switch event.eventType {
	case connect, reauth:
		if err := filterSubnets(m.commandWriter, event.clientID, m.filter); err != nil {
			log.Error("Unable to authenticate client:", err)
		}
	}
}

func filterSubnets(commandWriter management.CommandWriter, clientID int, subnets []string) error {
	data := struct {
		ClientID int
		Subnets  []string
	}{ClientID: clientID, Subnets: subnets}

	var tpl bytes.Buffer
	if err := filterLAN.Execute(&tpl, data); err != nil {
		return err
	}

	_, err := commandWriter.SingleLineCommand(tpl.String())
	return err
}
