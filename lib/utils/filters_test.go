package utils

import (
	"reflect"
	"testing"
)

func TestGetTopicInfo(t *testing.T) {
	type args struct {
		topicFilter string
	}
	tests := []struct {
		name          string
		args          args
		wantLevels    []string
		wantIsShared  bool
		wantShareName string
		wantErr       bool
	}{
		// TODO: Add test cases.
		{
			"Single level wildcard",
			args{
				"sport/+/player1",
			},
			[]string{"sport", "+", "player1"},
			false,
			"",
			false,
		},
		{
			"Single level wildcard ending with multi level",
			args{
				"+/tennis/#",
			},
			[]string{"+", "tennis", "#"},
			false,
			"",
			false,
		},
		{
			"Shared subscription ending with multi level",
			args{
				"$share/consumer1/sports/tennis/#",
			},
			[]string{"sports", "tennis", "#"},
			true,
			"consumer1",
			false,
		},
		{
			"Invalid shared subscription",
			args{
				"$share/consumer1",
			},
			nil,
			false,
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLevels, gotIsShared, gotShareName, err := GetTopicInfo(tt.args.topicFilter)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTopicInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotLevels, tt.wantLevels) {
				t.Errorf("GetTopicInfo() gotLevels = %v, want %v", gotLevels, tt.wantLevels)
			}
			if gotIsShared != tt.wantIsShared {
				t.Errorf("GetTopicInfo() gotIsShared = %v, want %v", gotIsShared, tt.wantIsShared)
			}
			if gotShareName != tt.wantShareName {
				t.Errorf("GetTopicInfo() gotShareName = %v, want %v", gotShareName, tt.wantShareName)
			}
		})
	}
}

func TestTopicMatches(t *testing.T) {
	type args struct {
		topic       string
		topicFilter string
	}
	tests := []struct {
		name          string
		args          args
		wantMatches   bool
		wantIsShared  bool
		wantShareName string
	}{
		// TODO: Add test cases.
		{
			"Test 1",
			args{
				"sport/tennis/player1",
				"sport/tennis/player1/#",
			},
			true,
			false,
			"",
		},
		{
			"Test 2",
			args{
				"sport/tennis/player1/ranking",
				"sport/tennis/player1/#",
			},
			true,
			false,
			"",
		},
		{
			"Test 3",
			args{
				"sport/tennis/player1/score/wimbledon",
				"sport/tennis/player1/#",
			},
			true,
			false,
			"",
		},
		{
			"Test 4",
			args{
				"sport",
				"sport/#",
			},
			true,
			false,
			"",
		},
		{
			"Test 5",
			args{
				"sport/tennis/player1",
				"sport/+/player1",
			},
			true,
			false,
			"",
		},
		{
			"Test 6",
			args{
				"sport/tennis/player1",
				"$share/consumer/sport/+/player1",
			},
			true,
			true,
			"consumer",
		},
		{
			"Test 7",
			args{
				"sport/tennis/player1/tournaments/schedule",
				"$share/consumer/sport/+/+/#",
			},
			true,
			true,
			"consumer",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatches, gotIsShared, gotShareName := TopicMatches(tt.args.topic, tt.args.topicFilter)
			if gotMatches != tt.wantMatches {
				t.Errorf("TopicMatches() gotMatches = %v, want %v", gotMatches, tt.wantMatches)
			}
			if gotIsShared != tt.wantIsShared {
				t.Errorf("TopicMatches() gotIsShared = %v, want %v", gotIsShared, tt.wantIsShared)
			}
			if gotShareName != tt.wantShareName {
				t.Errorf("TopicMatches() gotShareName = %v, want %v", gotShareName, tt.wantShareName)
			}
		})
	}
}
