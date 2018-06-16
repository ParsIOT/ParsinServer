// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package glb

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

func easyjsonF5bcc766DecodeParsinServerGlb(in *jlexer.Lexer, out *UserPositionJSON) {
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
		case "time":
			out.Time = int64(in.Int64())
		case "location":
			out.Location = string(in.String())
		case "bayesguess":
			out.BayesGuess = string(in.String())
		case "bayesdata":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.BayesData = make(map[string]float64)
				} else {
					out.BayesData = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v1 float64
					v1 = float64(in.Float64())
					(out.BayesData)[key] = v1
					in.WantComma()
				}
				in.Delim('}')
			}
		case "svmguess":
			out.SvmGuess = string(in.String())
		case "svmdata":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.SvmData = make(map[string]float64)
				} else {
					out.SvmData = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v2 float64
					v2 = float64(in.Float64())
					(out.SvmData)[key] = v2
					in.WantComma()
				}
				in.Delim('}')
			}
		case "rfdata":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.ScikitData = make(map[string]string)
				} else {
					out.ScikitData = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v3 string
					v3 = string(in.String())
					(out.ScikitData)[key] = v3
					in.WantComma()
				}
				in.Delim('}')
			}
		case "knnguess":
			out.KnnGuess = string(in.String())
		case "knndata":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.KnnData = make(map[string]float64)
				} else {
					out.KnnData = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v4 float64
					v4 = float64(in.Float64())
					(out.KnnData)[key] = v4
					in.WantComma()
				}
				in.Delim('}')
			}
		case "pdrlocation":
			out.PDRLocation = string(in.String())
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
func easyjsonF5bcc766EncodeParsinServerGlb(out *jwriter.Writer, in UserPositionJSON) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"time\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64(int64(in.Time))
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
		const prefix string = ",\"bayesguess\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.BayesGuess))
	}
	{
		const prefix string = ",\"bayesdata\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.BayesData == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v5First := true
			for v5Name, v5Value := range in.BayesData {
				if v5First {
					v5First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v5Name))
				out.RawByte(':')
				out.Float64(float64(v5Value))
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"svmguess\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.SvmGuess))
	}
	{
		const prefix string = ",\"svmdata\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.SvmData == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v6First := true
			for v6Name, v6Value := range in.SvmData {
				if v6First {
					v6First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v6Name))
				out.RawByte(':')
				out.Float64(float64(v6Value))
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"rfdata\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.ScikitData == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v7First := true
			for v7Name, v7Value := range in.ScikitData {
				if v7First {
					v7First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v7Name))
				out.RawByte(':')
				out.String(string(v7Value))
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"knnguess\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.KnnGuess))
	}
	{
		const prefix string = ",\"knndata\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.KnnData == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v8First := true
			for v8Name, v8Value := range in.KnnData {
				if v8First {
					v8First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v8Name))
				out.RawByte(':')
				out.Float64(float64(v8Value))
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"pdrlocation\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.PDRLocation))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v UserPositionJSON) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonF5bcc766EncodeParsinServerGlb(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v UserPositionJSON) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonF5bcc766EncodeParsinServerGlb(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *UserPositionJSON) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonF5bcc766DecodeParsinServerGlb(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *UserPositionJSON) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonF5bcc766DecodeParsinServerGlb(l, v)
}
func easyjsonF5bcc766DecodeParsinServerGlb1(in *jlexer.Lexer, out *FilterMacs) {
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
		case "macs":
			if in.IsNull() {
				in.Skip()
				out.Macs = nil
			} else {
				in.Delim('[')
				if out.Macs == nil {
					if !in.IsDelim(']') {
						out.Macs = make([]string, 0, 4)
					} else {
						out.Macs = []string{}
					}
				} else {
					out.Macs = (out.Macs)[:0]
				}
				for !in.IsDelim(']') {
					var v9 string
					v9 = string(in.String())
					out.Macs = append(out.Macs, v9)
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
func easyjsonF5bcc766EncodeParsinServerGlb1(out *jwriter.Writer, in FilterMacs) {
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
		const prefix string = ",\"macs\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Macs == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v10, v11 := range in.Macs {
				if v10 > 0 {
					out.RawByte(',')
				}
				out.String(string(v11))
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v FilterMacs) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonF5bcc766EncodeParsinServerGlb1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v FilterMacs) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonF5bcc766EncodeParsinServerGlb1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *FilterMacs) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonF5bcc766DecodeParsinServerGlb1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *FilterMacs) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonF5bcc766DecodeParsinServerGlb1(l, v)
}
func easyjsonF5bcc766DecodeParsinServerGlb2(in *jlexer.Lexer, out *Empty) {
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
func easyjsonF5bcc766EncodeParsinServerGlb2(out *jwriter.Writer, in Empty) {
	out.RawByte('{')
	first := true
	_ = first
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Empty) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonF5bcc766EncodeParsinServerGlb2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Empty) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonF5bcc766EncodeParsinServerGlb2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Empty) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonF5bcc766DecodeParsinServerGlb2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Empty) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonF5bcc766DecodeParsinServerGlb2(l, v)
}