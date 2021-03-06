/*
 * Copyright 2018-present Open Networking Foundation

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package kafka

import (
	"fmt"
	"strings"
)

const (
	TopicSeparator = "_"
	DeviceIdLength = 24
)

// A Topic definition - may be augmented with additional attributes eventually
type Topic struct {
	// The name of the topic. It must start with a letter,
	// and contain only letters (`[A-Za-z]`), numbers (`[0-9]`), dashes (`-`),
	// underscores (`_`), periods (`.`), tildes (`~`), plus (`+`) or percent
	// signs (`%`).
	Name string
}

type KVArg struct {
	Key   string
	Value interface{}
}

//CreateSubTopic concatenate a list of arguments together using underscores.
func CreateSubTopic(args ...string) Topic {
	topic := ""
	for index, arg := range args {
		if index == 0 {
			topic = arg
		} else {
			topic = fmt.Sprintf("%s%s%s", topic, TopicSeparator, arg)
		}
	}
	return Topic{Name: topic}
}

// GetDeviceIdFromTopic extract the deviceId from the topic name.  The topic name is formatted either as:
//			<any string> or <any string>_<deviceId>.  The device Id is 24 characters long.
func GetDeviceIdFromTopic(topic Topic) string {
	pos := strings.LastIndex(topic.Name, TopicSeparator)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(TopicSeparator)
	if adjustedPos >= len(topic.Name) {
		return ""
	}
	deviceId := topic.Name[adjustedPos:len(topic.Name)]
	if len(deviceId) != DeviceIdLength {
		return ""
	}
	return deviceId
}
