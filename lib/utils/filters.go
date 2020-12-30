package utils

import (
	"errors"
	"strings"
)

func GetTopicInfo(topicFilter string) (levels []string, isShared bool, shareName string, err error) {
	levels = strings.Split(topicFilter, "/")

	if strings.EqualFold("$share", levels[0]) {
		if len(levels) >= 3 {
			return levels[2:], true, levels[1], nil
		} else {
			return nil, false, "", errors.New("invalid shared subscription")
		}
	}

	return levels, false, "", nil
}

func TopicMatches(topic string, topicFilter string) (matches bool, isShared bool, shareName string) {
	levels, isShared, shareName, err := GetTopicInfo(topicFilter)
	if err != nil {
		return false, false, ""
	}

	incomingTopicChunks := strings.Split(topic, "/")

	for index, level := range levels {
		if strings.EqualFold(level, "#") {
			return true, isShared, shareName
		}
		if strings.EqualFold(level, "+") {
			continue
		} else {
			if !strings.EqualFold(level, incomingTopicChunks[index]) {
				return false, false, ""
			}
		}
	}

	return true, isShared, shareName
}
