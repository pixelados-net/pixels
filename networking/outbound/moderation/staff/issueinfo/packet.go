// Package issueinfo contains ISSUE_INFO projection.
package issueinfo

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ISSUE_INFO.
const Header uint16 = 3609

// Pattern stores one highlighted text match.
type Pattern struct {
	Pattern string
	Start   int32
	End     int32
}

// Params stores one moderation issue summary.
type Params struct {
	IssueID            int32
	State              int32
	CategoryID         int32
	ReportedCategoryID int32
	AgeMilliseconds    int32
	Priority           int32
	GroupingID         int32
	ReporterID         int32
	ReporterName       string
	ReportedID         int32
	ReportedName       string
	PickerID           int32
	PickerName         string
	Message            string
	ChatRecordID       int32
	Patterns           []Pattern
}

// Encode creates the exact Nitro issue summary.
func Encode(params Params) (codec.Packet, error) {
	definition := codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.StringField, codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field}
	payload, err := codec.AppendPayload(nil, definition, codec.Int32(params.IssueID), codec.Int32(params.State), codec.Int32(params.CategoryID), codec.Int32(params.ReportedCategoryID), codec.Int32(params.AgeMilliseconds), codec.Int32(params.Priority), codec.Int32(params.GroupingID), codec.Int32(params.ReporterID), codec.String(params.ReporterName), codec.Int32(params.ReportedID), codec.String(params.ReportedName), codec.Int32(params.PickerID), codec.String(params.PickerName), codec.String(params.Message), codec.Int32(params.ChatRecordID), codec.Int32(int32(len(params.Patterns))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, pattern := range params.Patterns {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field}, codec.String(pattern.Pattern), codec.Int32(pattern.Start), codec.Int32(pattern.End))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}
