package main

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	testCases := []struct {
		lamport bool
		want    []Event
	}{
		{
			lamport: false,
			want: []Event{
				{
					Clock:      1,
					Message:    Message{AbsoluteId: 1, Clock: 11, SenderId: 0},
					ReceiverId: 2,
				},
				{
					Clock:      1,
					Message:    Message{AbsoluteId: 2, Clock: 2, SenderId: 2},
					ReceiverId: 1,
				},
				{
					Clock:      12,
					Message:    Message{AbsoluteId: 3, Clock: 2, SenderId: 1},
					ReceiverId: 0,
				},
			},
		},
		{
			lamport: true,
			want: []Event{
				{
					Clock:      12,
					Message:    Message{AbsoluteId: 1, Clock: 11, SenderId: 0},
					ReceiverId: 2,
				},
				{
					Clock:      14,
					Message:    Message{AbsoluteId: 2, Clock: 13, SenderId: 2},
					ReceiverId: 1,
				},
				{
					Clock:      16,
					Message:    Message{AbsoluteId: 3, Clock: 15, SenderId: 1},
					ReceiverId: 0,
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run("lamport: "+strconv.FormatBool(tC.lamport), func(t *testing.T) {
			got, err := Run(tC.lamport)
			assert.NoError(t, err)
			assert.Equal(t, tC.want, got)
		})
	}
}
