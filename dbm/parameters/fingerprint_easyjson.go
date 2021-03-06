// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package parameters

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson7b60966aDecodeParsinServerDbmParameters(in *jlexer.Lexer, out *Router) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "mac":
			out.Mac = string(in.String())
		case "rssi":
			out.Rssi = int(in.Int())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson7b60966aEncodeParsinServerDbmParameters(out *jwriter.Writer, in Router) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"mac\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Mac))
	}
	{
		const prefix string = ",\"rssi\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.Rssi))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Router) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson7b60966aEncodeParsinServerDbmParameters(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Router) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson7b60966aEncodeParsinServerDbmParameters(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Router) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson7b60966aDecodeParsinServerDbmParameters(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Router) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson7b60966aDecodeParsinServerDbmParameters(l, v)
}
func easyjson7b60966aDecodeParsinServerDbmParameters1(in *jlexer.Lexer, out *Fingerprint) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "group":
			out.Group = string(in.String())
		case "username":
			out.Username = string(in.String())
		case "location":
			out.Location = string(in.String())
		case "time":
			out.Timestamp = int64(in.Int64())
		case "wifi-fingerprint":
			if in.IsNull() {
				in.Skip()
				out.WifiFingerprint = nil
			} else {
				in.Delim('[')
				if out.WifiFingerprint == nil {
					if !in.IsDelim(']') {
						out.WifiFingerprint = make([]Router, 0, 2)
					} else {
						out.WifiFingerprint = []Router{}
					}
				} else {
					out.WifiFingerprint = (out.WifiFingerprint)[:0]
				}
				for !in.IsDelim(']') {
					var v1 Router
					(v1).UnmarshalEasyJSON(in)
					out.WifiFingerprint = append(out.WifiFingerprint, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "test-validation":
			out.TestValidation = bool(in.Bool())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson7b60966aEncodeParsinServerDbmParameters1(out *jwriter.Writer, in Fingerprint) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"group\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Group))
	}
	{
		const prefix string = ",\"username\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Username))
	}
	{
		const prefix string = ",\"location\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Location))
	}
	{
		const prefix string = ",\"time\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64(int64(in.Timestamp))
	}
	{
		const prefix string = ",\"wifi-fingerprint\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.WifiFingerprint == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.WifiFingerprint {
				if v2 > 0 {
					out.RawByte(',')
				}
				(v3).MarshalEasyJSON(out)
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"test-validation\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Bool(bool(in.TestValidation))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Fingerprint) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson7b60966aEncodeParsinServerDbmParameters1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Fingerprint) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson7b60966aEncodeParsinServerDbmParameters1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Fingerprint) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson7b60966aDecodeParsinServerDbmParameters1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Fingerprint) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson7b60966aDecodeParsinServerDbmParameters1(l, v)
}
func easyjson7b60966aDecodeParsinServerDbmParameters2(in *jlexer.Lexer, out *BulkFingerprint) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "fingerprints":
			if in.IsNull() {
				in.Skip()
				out.Fingerprints = nil
			} else {
				in.Delim('[')
				if out.Fingerprints == nil {
					if !in.IsDelim(']') {
						out.Fingerprints = make([]Fingerprint, 0, 1)
					} else {
						out.Fingerprints = []Fingerprint{}
					}
				} else {
					out.Fingerprints = (out.Fingerprints)[:0]
				}
				for !in.IsDelim(']') {
					var v4 Fingerprint
					(v4).UnmarshalEasyJSON(in)
					out.Fingerprints = append(out.Fingerprints, v4)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson7b60966aEncodeParsinServerDbmParameters2(out *jwriter.Writer, in BulkFingerprint) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"fingerprints\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Fingerprints == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v5, v6 := range in.Fingerprints {
				if v5 > 0 {
					out.RawByte(',')
				}
				(v6).MarshalEasyJSON(out)
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v BulkFingerprint) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson7b60966aEncodeParsinServerDbmParameters2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v BulkFingerprint) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson7b60966aEncodeParsinServerDbmParameters2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *BulkFingerprint) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson7b60966aDecodeParsinServerDbmParameters2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *BulkFingerprint) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson7b60966aDecodeParsinServerDbmParameters2(l, v)
}
