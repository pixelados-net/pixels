// Package contents encodes POLL_CONTENTS responses.
package contents

import "github.com/niflaot/pixels/networking/codec"

// Header identifies POLL_CONTENTS.
const Header uint16 = 2997

// Choice describes one selectable poll answer.
type Choice struct {
	// Value stores the submitted value.
	Value string
	// Text stores the visible label.
	Text string
	// Type stores the client choice type.
	Type int32
}

// Question describes one poll question or conditional child.
type Question struct {
	// ID identifies the question.
	ID int32
	// SortOrder stores display order.
	SortOrder int32
	// Type stores free-text, single-choice, or multi-choice behavior.
	Type int32
	// Text stores visible question text.
	Text string
	// Category stores the analytics category.
	Category int32
	// AnswerType stores the expected answer representation.
	AnswerType int32
	// Choices stores selectable answers for types one and two.
	Choices []Choice
	// Children stores conditional child questions.
	Children []Question
}

// Data describes one complete DB-backed poll.
type Data struct {
	// ID identifies the poll.
	ID int32
	// StartMessage stores introductory text.
	StartMessage string
	// EndMessage stores completion text.
	EndMessage string
	// Questions stores top-level questions.
	Questions []Question
	// NPS reports whether the poll is a net-promoter survey.
	NPS bool
}

// Encode creates one complete poll content response.
func Encode(data Data) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field}, codec.Int32(data.ID), codec.String(data.StartMessage), codec.String(data.EndMessage), codec.Int32(int32(len(data.Questions))))
	for _, question := range data.Questions {
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = appendQuestion(payload, question)
		if err != nil {
			return codec.Packet{}, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(question.Children))))
		for _, child := range question.Children {
			if err != nil {
				return codec.Packet{}, err
			}
			payload, err = appendQuestion(payload, child)
		}
	}
	if err != nil {
		return codec.Packet{}, err
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.BooleanField}, codec.Bool(data.NPS))
	return codec.Packet{Header: Header, Payload: payload}, err
}

// appendQuestion appends one renderer poll-question shape.
func appendQuestion(payload []byte, question Question) ([]byte, error) {
	payload, err := codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(question.ID), codec.Int32(question.SortOrder), codec.Int32(question.Type), codec.String(question.Text), codec.Int32(question.Category), codec.Int32(question.AnswerType), codec.Int32(int32(len(question.Choices))))
	if question.Type != 1 && question.Type != 2 {
		return payload, err
	}
	for _, choice := range question.Choices {
		if err != nil {
			return nil, err
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.StringField, codec.Int32Field}, codec.String(choice.Value), codec.String(choice.Text), codec.Int32(choice.Type))
	}
	return payload, err
}
