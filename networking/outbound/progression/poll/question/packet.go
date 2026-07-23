// Package question encodes QUESTION responses.
package question

import "github.com/niflaot/pixels/networking/codec"

// Header identifies QUESTION.
const Header uint16 = 2665

// Selection describes one selectable answer.
type Selection struct {
	// Value stores the submitted wire value.
	Value string
	// Label stores the localized display label.
	Label string
}

// Data describes one room poll question.
type Data struct {
	// PollType stores the client widget type.
	PollType string
	// PollID identifies the poll.
	PollID int32
	// QuestionID identifies the active question.
	QuestionID int32
	// DurationSeconds stores the answer window.
	DurationSeconds int32
	// ID identifies embedded question content.
	ID int32
	// Number stores the display order.
	Number int32
	// Type stores word, single, or multiple behavior.
	Type int32
	// Content stores the question text.
	Content string
	// SelectionMinimum stores the minimum choice count.
	SelectionMinimum int32
	// Selections stores optional choices.
	Selections []Selection
}

// Encode creates one room poll question response.
func Encode(data Data) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField}, codec.String(data.PollType), codec.Int32(data.PollID), codec.Int32(data.QuestionID), codec.Int32(data.DurationSeconds), codec.Int32(data.ID), codec.Int32(data.Number), codec.Int32(data.Type), codec.String(data.Content))
	if (data.Type == 1 || data.Type == 2) && err == nil {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(data.SelectionMinimum), codec.Int32(int32(len(data.Selections))))
		for _, selection := range data.Selections {
			if err != nil {
				return codec.Packet{}, err
			}
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.StringField}, codec.String(selection.Value), codec.String(selection.Label))
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
