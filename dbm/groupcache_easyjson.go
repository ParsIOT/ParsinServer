// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package dbm

import (
	parameters "ParsinServer/algorithms/parameters"
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

func easyjson3b8810b5DecodeParsinServerDbm(in *jlexer.Lexer, out *ResultDataStruct) {
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
		case "Results":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.Results = make(map[string]parameters.Fingerprint)
				} else {
					out.Results = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v1 parameters.Fingerprint
					(v1).UnmarshalEasyJSON(in)
					(out.Results)[key] = v1
					in.WantComma()
				}
				in.Delim('}')
			}
		case "AlgoAccuracy":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.AlgoAccuracy = make(map[string]int)
				} else {
					out.AlgoAccuracy = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v2 int
					v2 = int(in.Int())
					(out.AlgoAccuracy)[key] = v2
					in.WantComma()
				}
				in.Delim('}')
			}
		case "AlgoAccuracyLoc":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.AlgoAccuracyLoc = make(map[string]map[string]int)
				} else {
					out.AlgoAccuracyLoc = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v3 map[string]int
					if in.IsNull() {
						in.Skip()
					} else {
						in.Delim('{')
						if !in.IsDelim('}') {
							v3 = make(map[string]int)
						} else {
							v3 = nil
						}
						for !in.IsDelim('}') {
							key := string(in.String())
							in.WantColon()
							var v4 int
							v4 = int(in.Int())
							(v3)[key] = v4
							in.WantComma()
						}
						in.Delim('}')
					}
					(out.AlgoAccuracyLoc)[key] = v3
					in.WantComma()
				}
				in.Delim('}')
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
func easyjson3b8810b5EncodeParsinServerDbm(out *jwriter.Writer, in ResultDataStruct) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"Results\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Results == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v5First := true
			for v5Name, v5Value := range in.Results {
				if v5First {
					v5First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v5Name))
				out.RawByte(':')
				(v5Value).MarshalEasyJSON(out)
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"AlgoAccuracy\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.AlgoAccuracy == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v6First := true
			for v6Name, v6Value := range in.AlgoAccuracy {
				if v6First {
					v6First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v6Name))
				out.RawByte(':')
				out.Int(int(v6Value))
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"AlgoAccuracyLoc\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.AlgoAccuracyLoc == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v7First := true
			for v7Name, v7Value := range in.AlgoAccuracyLoc {
				if v7First {
					v7First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v7Name))
				out.RawByte(':')
				if v7Value == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
					out.RawString(`null`)
				} else {
					out.RawByte('{')
					v8First := true
					for v8Name, v8Value := range v7Value {
						if v8First {
							v8First = false
						} else {
							out.RawByte(',')
						}
						out.String(string(v8Name))
						out.RawByte(':')
						out.Int(int(v8Value))
					}
					out.RawByte('}')
				}
			}
			out.RawByte('}')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ResultDataStruct) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson3b8810b5EncodeParsinServerDbm(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ResultDataStruct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson3b8810b5EncodeParsinServerDbm(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ResultDataStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3b8810b5DecodeParsinServerDbm(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ResultDataStruct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson3b8810b5DecodeParsinServerDbm(l, v)
}
func easyjson3b8810b5DecodeParsinServerDbm1(in *jlexer.Lexer, out *RawDataStruct) {
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
		case "Fingerprints":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.Fingerprints = make(map[string]parameters.Fingerprint)
				} else {
					out.Fingerprints = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v9 parameters.Fingerprint
					(v9).UnmarshalEasyJSON(in)
					(out.Fingerprints)[key] = v9
					in.WantComma()
				}
				in.Delim('}')
			}
		case "FingerprintsOrdering":
			if in.IsNull() {
				in.Skip()
				out.FingerprintsOrdering = nil
			} else {
				in.Delim('[')
				if out.FingerprintsOrdering == nil {
					if !in.IsDelim(']') {
						out.FingerprintsOrdering = make([]string, 0, 4)
					} else {
						out.FingerprintsOrdering = []string{}
					}
				} else {
					out.FingerprintsOrdering = (out.FingerprintsOrdering)[:0]
				}
				for !in.IsDelim(']') {
					var v10 string
					v10 = string(in.String())
					out.FingerprintsOrdering = append(out.FingerprintsOrdering, v10)
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
func easyjson3b8810b5EncodeParsinServerDbm1(out *jwriter.Writer, in RawDataStruct) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"Fingerprints\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Fingerprints == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v11First := true
			for v11Name, v11Value := range in.Fingerprints {
				if v11First {
					v11First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v11Name))
				out.RawByte(':')
				(v11Value).MarshalEasyJSON(out)
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"FingerprintsOrdering\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.FingerprintsOrdering == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v12, v13 := range in.FingerprintsOrdering {
				if v12 > 0 {
					out.RawByte(',')
				}
				out.String(string(v13))
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v RawDataStruct) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson3b8810b5EncodeParsinServerDbm1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v RawDataStruct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson3b8810b5EncodeParsinServerDbm1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *RawDataStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3b8810b5DecodeParsinServerDbm1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *RawDataStruct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson3b8810b5DecodeParsinServerDbm1(l, v)
}
func easyjson3b8810b5DecodeParsinServerDbm2(in *jlexer.Lexer, out *MiddleDataStruct) {
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
		case "NetworkMacs":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.NetworkMacs = make(map[string]map[string]bool)
				} else {
					out.NetworkMacs = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v14 map[string]bool
					if in.IsNull() {
						in.Skip()
					} else {
						in.Delim('{')
						if !in.IsDelim('}') {
							v14 = make(map[string]bool)
						} else {
							v14 = nil
						}
						for !in.IsDelim('}') {
							key := string(in.String())
							in.WantColon()
							var v15 bool
							v15 = bool(in.Bool())
							(v14)[key] = v15
							in.WantComma()
						}
						in.Delim('}')
					}
					(out.NetworkMacs)[key] = v14
					in.WantComma()
				}
				in.Delim('}')
			}
		case "NetworkLocs":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.NetworkLocs = make(map[string]map[string]bool)
				} else {
					out.NetworkLocs = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v16 map[string]bool
					if in.IsNull() {
						in.Skip()
					} else {
						in.Delim('{')
						if !in.IsDelim('}') {
							v16 = make(map[string]bool)
						} else {
							v16 = nil
						}
						for !in.IsDelim('}') {
							key := string(in.String())
							in.WantColon()
							var v17 bool
							v17 = bool(in.Bool())
							(v16)[key] = v17
							in.WantComma()
						}
						in.Delim('}')
					}
					(out.NetworkLocs)[key] = v16
					in.WantComma()
				}
				in.Delim('}')
			}
		case "MacVariability":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.MacVariability = make(map[string]float32)
				} else {
					out.MacVariability = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v18 float32
					v18 = float32(in.Float32())
					(out.MacVariability)[key] = v18
					in.WantComma()
				}
				in.Delim('}')
			}
		case "MacCount":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.MacCount = make(map[string]int)
				} else {
					out.MacCount = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v19 int
					v19 = int(in.Int())
					(out.MacCount)[key] = v19
					in.WantComma()
				}
				in.Delim('}')
			}
		case "MacCountByLoc":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.MacCountByLoc = make(map[string]map[string]int)
				} else {
					out.MacCountByLoc = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v20 map[string]int
					if in.IsNull() {
						in.Skip()
					} else {
						in.Delim('{')
						if !in.IsDelim('}') {
							v20 = make(map[string]int)
						} else {
							v20 = nil
						}
						for !in.IsDelim('}') {
							key := string(in.String())
							in.WantColon()
							var v21 int
							v21 = int(in.Int())
							(v20)[key] = v21
							in.WantComma()
						}
						in.Delim('}')
					}
					(out.MacCountByLoc)[key] = v20
					in.WantComma()
				}
				in.Delim('}')
			}
		case "UniqueLocs":
			if in.IsNull() {
				in.Skip()
				out.UniqueLocs = nil
			} else {
				in.Delim('[')
				if out.UniqueLocs == nil {
					if !in.IsDelim(']') {
						out.UniqueLocs = make([]string, 0, 4)
					} else {
						out.UniqueLocs = []string{}
					}
				} else {
					out.UniqueLocs = (out.UniqueLocs)[:0]
				}
				for !in.IsDelim(']') {
					var v22 string
					v22 = string(in.String())
					out.UniqueLocs = append(out.UniqueLocs, v22)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "UniqueMacs":
			if in.IsNull() {
				in.Skip()
				out.UniqueMacs = nil
			} else {
				in.Delim('[')
				if out.UniqueMacs == nil {
					if !in.IsDelim(']') {
						out.UniqueMacs = make([]string, 0, 4)
					} else {
						out.UniqueMacs = []string{}
					}
				} else {
					out.UniqueMacs = (out.UniqueMacs)[:0]
				}
				for !in.IsDelim(']') {
					var v23 string
					v23 = string(in.String())
					out.UniqueMacs = append(out.UniqueMacs, v23)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "LocCount":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.LocCount = make(map[string]int)
				} else {
					out.LocCount = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v24 int
					v24 = int(in.Int())
					(out.LocCount)[key] = v24
					in.WantComma()
				}
				in.Delim('}')
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
func easyjson3b8810b5EncodeParsinServerDbm2(out *jwriter.Writer, in MiddleDataStruct) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"NetworkMacs\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.NetworkMacs == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v25First := true
			for v25Name, v25Value := range in.NetworkMacs {
				if v25First {
					v25First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v25Name))
				out.RawByte(':')
				if v25Value == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
					out.RawString(`null`)
				} else {
					out.RawByte('{')
					v26First := true
					for v26Name, v26Value := range v25Value {
						if v26First {
							v26First = false
						} else {
							out.RawByte(',')
						}
						out.String(string(v26Name))
						out.RawByte(':')
						out.Bool(bool(v26Value))
					}
					out.RawByte('}')
				}
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"NetworkLocs\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.NetworkLocs == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v27First := true
			for v27Name, v27Value := range in.NetworkLocs {
				if v27First {
					v27First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v27Name))
				out.RawByte(':')
				if v27Value == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
					out.RawString(`null`)
				} else {
					out.RawByte('{')
					v28First := true
					for v28Name, v28Value := range v27Value {
						if v28First {
							v28First = false
						} else {
							out.RawByte(',')
						}
						out.String(string(v28Name))
						out.RawByte(':')
						out.Bool(bool(v28Value))
					}
					out.RawByte('}')
				}
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"MacVariability\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.MacVariability == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v29First := true
			for v29Name, v29Value := range in.MacVariability {
				if v29First {
					v29First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v29Name))
				out.RawByte(':')
				out.Float32(float32(v29Value))
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"MacCount\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.MacCount == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v30First := true
			for v30Name, v30Value := range in.MacCount {
				if v30First {
					v30First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v30Name))
				out.RawByte(':')
				out.Int(int(v30Value))
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"MacCountByLoc\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.MacCountByLoc == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v31First := true
			for v31Name, v31Value := range in.MacCountByLoc {
				if v31First {
					v31First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v31Name))
				out.RawByte(':')
				if v31Value == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
					out.RawString(`null`)
				} else {
					out.RawByte('{')
					v32First := true
					for v32Name, v32Value := range v31Value {
						if v32First {
							v32First = false
						} else {
							out.RawByte(',')
						}
						out.String(string(v32Name))
						out.RawByte(':')
						out.Int(int(v32Value))
					}
					out.RawByte('}')
				}
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"UniqueLocs\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.UniqueLocs == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v33, v34 := range in.UniqueLocs {
				if v33 > 0 {
					out.RawByte(',')
				}
				out.String(string(v34))
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"UniqueMacs\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.UniqueMacs == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v35, v36 := range in.UniqueMacs {
				if v35 > 0 {
					out.RawByte(',')
				}
				out.String(string(v36))
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"LocCount\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.LocCount == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v37First := true
			for v37Name, v37Value := range in.LocCount {
				if v37First {
					v37First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v37Name))
				out.RawByte(':')
				out.Int(int(v37Value))
			}
			out.RawByte('}')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v MiddleDataStruct) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson3b8810b5EncodeParsinServerDbm2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v MiddleDataStruct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson3b8810b5EncodeParsinServerDbm2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *MiddleDataStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3b8810b5DecodeParsinServerDbm2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *MiddleDataStruct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson3b8810b5DecodeParsinServerDbm2(l, v)
}
func easyjson3b8810b5DecodeParsinServerDbm3(in *jlexer.Lexer, out *GroupManger) {
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
func easyjson3b8810b5EncodeParsinServerDbm3(out *jwriter.Writer, in GroupManger) {
	out.RawByte('{')
	first := true
	_ = first
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v GroupManger) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson3b8810b5EncodeParsinServerDbm3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v GroupManger) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson3b8810b5EncodeParsinServerDbm3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *GroupManger) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3b8810b5DecodeParsinServerDbm3(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *GroupManger) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson3b8810b5DecodeParsinServerDbm3(l, v)
}
func easyjson3b8810b5DecodeParsinServerDbm4(in *jlexer.Lexer, out *Group) {
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
		case "Name":
			out.Name = string(in.String())
		case "Permanent":
			out.Permanent = bool(in.Bool())
		case "RawData":
			if in.IsNull() {
				in.Skip()
				out.RawData = nil
			} else {
				if out.RawData == nil {
					out.RawData = new(RawDataStruct)
				}
				(*out.RawData).UnmarshalEasyJSON(in)
			}
		case "MiddleData":
			if in.IsNull() {
				in.Skip()
				out.MiddleData = nil
			} else {
				if out.MiddleData == nil {
					out.MiddleData = new(MiddleDataStruct)
				}
				(*out.MiddleData).UnmarshalEasyJSON(in)
			}
		case "AlgoData":
			if in.IsNull() {
				in.Skip()
				out.AlgoData = nil
			} else {
				if out.AlgoData == nil {
					out.AlgoData = new(AlgoDataStruct)
				}
				(*out.AlgoData).UnmarshalEasyJSON(in)
			}
		case "ResultData":
			if in.IsNull() {
				in.Skip()
				out.ResultData = nil
			} else {
				if out.ResultData == nil {
					out.ResultData = new(ResultDataStruct)
				}
				(*out.ResultData).UnmarshalEasyJSON(in)
			}
		case "RawDataChanged":
			out.RawDataChanged = bool(in.Bool())
		case "MiddleDataChanged":
			out.MiddleDataChanged = bool(in.Bool())
		case "AlgoDataChanged":
			out.AlgoDataChanged = bool(in.Bool())
		case "ResultDataChanged":
			out.ResultDataChanged = bool(in.Bool())
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
func easyjson3b8810b5EncodeParsinServerDbm4(out *jwriter.Writer, in Group) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"Name\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Name))
	}
	{
		const prefix string = ",\"Permanent\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Bool(bool(in.Permanent))
	}
	{
		const prefix string = ",\"RawData\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.RawData == nil {
			out.RawString("null")
		} else {
			(*in.RawData).MarshalEasyJSON(out)
		}
	}
	{
		const prefix string = ",\"MiddleData\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.MiddleData == nil {
			out.RawString("null")
		} else {
			(*in.MiddleData).MarshalEasyJSON(out)
		}
	}
	{
		const prefix string = ",\"AlgoData\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.AlgoData == nil {
			out.RawString("null")
		} else {
			(*in.AlgoData).MarshalEasyJSON(out)
		}
	}
	{
		const prefix string = ",\"ResultData\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.ResultData == nil {
			out.RawString("null")
		} else {
			(*in.ResultData).MarshalEasyJSON(out)
		}
	}
	{
		const prefix string = ",\"RawDataChanged\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Bool(bool(in.RawDataChanged))
	}
	{
		const prefix string = ",\"MiddleDataChanged\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Bool(bool(in.MiddleDataChanged))
	}
	{
		const prefix string = ",\"AlgoDataChanged\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Bool(bool(in.AlgoDataChanged))
	}
	{
		const prefix string = ",\"ResultDataChanged\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Bool(bool(in.ResultDataChanged))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Group) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson3b8810b5EncodeParsinServerDbm4(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Group) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson3b8810b5EncodeParsinServerDbm4(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Group) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3b8810b5DecodeParsinServerDbm4(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Group) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson3b8810b5DecodeParsinServerDbm4(l, v)
}
func easyjson3b8810b5DecodeParsinServerDbm5(in *jlexer.Lexer, out *AlgoDataStruct) {
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
		case "KnnFPs":
			(out.KnnFPs).UnmarshalEasyJSON(in)
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
func easyjson3b8810b5EncodeParsinServerDbm5(out *jwriter.Writer, in AlgoDataStruct) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"KnnFPs\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		(in.KnnFPs).MarshalEasyJSON(out)
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v AlgoDataStruct) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson3b8810b5EncodeParsinServerDbm5(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v AlgoDataStruct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson3b8810b5EncodeParsinServerDbm5(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *AlgoDataStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3b8810b5DecodeParsinServerDbm5(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *AlgoDataStruct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson3b8810b5DecodeParsinServerDbm5(l, v)
}
