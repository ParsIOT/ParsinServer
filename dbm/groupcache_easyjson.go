// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package dbm

import (
	parameters "ParsinServer/dbm/parameters"
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
	sync "sync"
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
					var v1 int
					v1 = int(in.Int())
					(out.AlgoAccuracy)[key] = v1
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
					var v2 map[string]int
					if in.IsNull() {
						in.Skip()
					} else {
						in.Delim('{')
						if !in.IsDelim('}') {
							v2 = make(map[string]int)
						} else {
							v2 = nil
						}
						for !in.IsDelim('}') {
							key := string(in.String())
							in.WantColon()
							var v3 int
							v3 = int(in.Int())
							(v2)[key] = v3
							in.WantComma()
						}
						in.Delim('}')
					}
					(out.AlgoAccuracyLoc)[key] = v2
					in.WantComma()
				}
				in.Delim('}')
			}
		case "UserHistory":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.UserHistory = make(map[string][]parameters.UserPositionJSON)
				} else {
					out.UserHistory = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v4 []parameters.UserPositionJSON
					if in.IsNull() {
						in.Skip()
						v4 = nil
					} else {
						in.Delim('[')
						if v4 == nil {
							if !in.IsDelim(']') {
								v4 = make([]parameters.UserPositionJSON, 0, 1)
							} else {
								v4 = []parameters.UserPositionJSON{}
							}
						} else {
							v4 = (v4)[:0]
						}
						for !in.IsDelim(']') {
							var v5 parameters.UserPositionJSON
							(v5).UnmarshalEasyJSON(in)
							v4 = append(v4, v5)
							in.WantComma()
						}
						in.Delim(']')
					}
					(out.UserHistory)[key] = v4
					in.WantComma()
				}
				in.Delim('}')
			}
		case "UserResults":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.UserResults = make(map[string][]parameters.UserPositionJSON)
				} else {
					out.UserResults = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v6 []parameters.UserPositionJSON
					if in.IsNull() {
						in.Skip()
						v6 = nil
					} else {
						in.Delim('[')
						if v6 == nil {
							if !in.IsDelim(']') {
								v6 = make([]parameters.UserPositionJSON, 0, 1)
							} else {
								v6 = []parameters.UserPositionJSON{}
							}
						} else {
							v6 = (v6)[:0]
						}
						for !in.IsDelim(']') {
							var v7 parameters.UserPositionJSON
							(v7).UnmarshalEasyJSON(in)
							v6 = append(v6, v7)
							in.WantComma()
						}
						in.Delim(']')
					}
					(out.UserResults)[key] = v6
					in.WantComma()
				}
				in.Delim('}')
			}
		case "AlgoTestErrorAccuracy":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.AlgoTestErrorAccuracy = make(map[string]int)
				} else {
					out.AlgoTestErrorAccuracy = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v8 int
					v8 = int(in.Int())
					(out.AlgoTestErrorAccuracy)[key] = v8
					in.WantComma()
				}
				in.Delim('}')
			}
		case "TestValidTracks":
			if in.IsNull() {
				in.Skip()
				out.TestValidTracks = nil
			} else {
				in.Delim('[')
				if out.TestValidTracks == nil {
					if !in.IsDelim(']') {
						out.TestValidTracks = make([]parameters.TestValidTrack, 0, 1)
					} else {
						out.TestValidTracks = []parameters.TestValidTrack{}
					}
				} else {
					out.TestValidTracks = (out.TestValidTracks)[:0]
				}
				for !in.IsDelim(']') {
					var v9 parameters.TestValidTrack
					(v9).UnmarshalEasyJSON(in)
					out.TestValidTracks = append(out.TestValidTracks, v9)
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
func easyjson3b8810b5EncodeParsinServerDbm(out *jwriter.Writer, in ResultDataStruct) {
	out.RawByte('{')
	first := true
	_ = first
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
			v10First := true
			for v10Name, v10Value := range in.AlgoAccuracy {
				if v10First {
					v10First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v10Name))
				out.RawByte(':')
				out.Int(int(v10Value))
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
			v11First := true
			for v11Name, v11Value := range in.AlgoAccuracyLoc {
				if v11First {
					v11First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v11Name))
				out.RawByte(':')
				if v11Value == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
					out.RawString(`null`)
				} else {
					out.RawByte('{')
					v12First := true
					for v12Name, v12Value := range v11Value {
						if v12First {
							v12First = false
						} else {
							out.RawByte(',')
						}
						out.String(string(v12Name))
						out.RawByte(':')
						out.Int(int(v12Value))
					}
					out.RawByte('}')
				}
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"UserHistory\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.UserHistory == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v13First := true
			for v13Name, v13Value := range in.UserHistory {
				if v13First {
					v13First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v13Name))
				out.RawByte(':')
				if v13Value == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
					out.RawString("null")
				} else {
					out.RawByte('[')
					for v14, v15 := range v13Value {
						if v14 > 0 {
							out.RawByte(',')
						}
						(v15).MarshalEasyJSON(out)
					}
					out.RawByte(']')
				}
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"UserResults\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.UserResults == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v16First := true
			for v16Name, v16Value := range in.UserResults {
				if v16First {
					v16First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v16Name))
				out.RawByte(':')
				if v16Value == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
					out.RawString("null")
				} else {
					out.RawByte('[')
					for v17, v18 := range v16Value {
						if v17 > 0 {
							out.RawByte(',')
						}
						(v18).MarshalEasyJSON(out)
					}
					out.RawByte(']')
				}
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"AlgoTestErrorAccuracy\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.AlgoTestErrorAccuracy == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v19First := true
			for v19Name, v19Value := range in.AlgoTestErrorAccuracy {
				if v19First {
					v19First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v19Name))
				out.RawByte(':')
				out.Int(int(v19Value))
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"TestValidTracks\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.TestValidTracks == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v20, v21 := range in.TestValidTracks {
				if v20 > 0 {
					out.RawByte(',')
				}
				(v21).MarshalEasyJSON(out)
			}
			out.RawByte(']')
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
					var v22 parameters.Fingerprint
					(v22).UnmarshalEasyJSON(in)
					(out.Fingerprints)[key] = v22
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
					var v23 string
					v23 = string(in.String())
					out.FingerprintsOrdering = append(out.FingerprintsOrdering, v23)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "LearnTrueLocations":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.LearnTrueLocations = make(map[int64]string)
				} else {
					out.LearnTrueLocations = nil
				}
				for !in.IsDelim('}') {
					key := int64(in.Int64Str())
					in.WantColon()
					var v24 string
					v24 = string(in.String())
					(out.LearnTrueLocations)[key] = v24
					in.WantComma()
				}
				in.Delim('}')
			}
		case "TestValidTrueLocations":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.TestValidTrueLocations = make(map[int64]string)
				} else {
					out.TestValidTrueLocations = nil
				}
				for !in.IsDelim('}') {
					key := int64(in.Int64Str())
					in.WantColon()
					var v25 string
					v25 = string(in.String())
					(out.TestValidTrueLocations)[key] = v25
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
			v26First := true
			for v26Name, v26Value := range in.Fingerprints {
				if v26First {
					v26First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v26Name))
				out.RawByte(':')
				(v26Value).MarshalEasyJSON(out)
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
			for v27, v28 := range in.FingerprintsOrdering {
				if v27 > 0 {
					out.RawByte(',')
				}
				out.String(string(v28))
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"LearnTrueLocations\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.LearnTrueLocations == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v29First := true
			for v29Name, v29Value := range in.LearnTrueLocations {
				if v29First {
					v29First = false
				} else {
					out.RawByte(',')
				}
				out.Int64Str(int64(v29Name))
				out.RawByte(':')
				out.String(string(v29Value))
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"TestValidTrueLocations\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.TestValidTrueLocations == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v30First := true
			for v30Name, v30Value := range in.TestValidTrueLocations {
				if v30First {
					v30First = false
				} else {
					out.RawByte(',')
				}
				out.Int64Str(int64(v30Name))
				out.RawByte(':')
				out.String(string(v30Value))
			}
			out.RawByte('}')
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
					var v31 map[string]bool
					if in.IsNull() {
						in.Skip()
					} else {
						in.Delim('{')
						if !in.IsDelim('}') {
							v31 = make(map[string]bool)
						} else {
							v31 = nil
						}
						for !in.IsDelim('}') {
							key := string(in.String())
							in.WantColon()
							var v32 bool
							v32 = bool(in.Bool())
							(v31)[key] = v32
							in.WantComma()
						}
						in.Delim('}')
					}
					(out.NetworkMacs)[key] = v31
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
					var v33 map[string]bool
					if in.IsNull() {
						in.Skip()
					} else {
						in.Delim('{')
						if !in.IsDelim('}') {
							v33 = make(map[string]bool)
						} else {
							v33 = nil
						}
						for !in.IsDelim('}') {
							key := string(in.String())
							in.WantColon()
							var v34 bool
							v34 = bool(in.Bool())
							(v33)[key] = v34
							in.WantComma()
						}
						in.Delim('}')
					}
					(out.NetworkLocs)[key] = v33
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
					var v35 float32
					v35 = float32(in.Float32())
					(out.MacVariability)[key] = v35
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
					var v36 int
					v36 = int(in.Int())
					(out.MacCount)[key] = v36
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
					var v37 map[string]int
					if in.IsNull() {
						in.Skip()
					} else {
						in.Delim('{')
						if !in.IsDelim('}') {
							v37 = make(map[string]int)
						} else {
							v37 = nil
						}
						for !in.IsDelim('}') {
							key := string(in.String())
							in.WantColon()
							var v38 int
							v38 = int(in.Int())
							(v37)[key] = v38
							in.WantComma()
						}
						in.Delim('}')
					}
					(out.MacCountByLoc)[key] = v37
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
					var v39 string
					v39 = string(in.String())
					out.UniqueLocs = append(out.UniqueLocs, v39)
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
					var v40 string
					v40 = string(in.String())
					out.UniqueMacs = append(out.UniqueMacs, v40)
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
					var v41 int
					v41 = int(in.Int())
					(out.LocCount)[key] = v41
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
			v42First := true
			for v42Name, v42Value := range in.NetworkMacs {
				if v42First {
					v42First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v42Name))
				out.RawByte(':')
				if v42Value == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
					out.RawString(`null`)
				} else {
					out.RawByte('{')
					v43First := true
					for v43Name, v43Value := range v42Value {
						if v43First {
							v43First = false
						} else {
							out.RawByte(',')
						}
						out.String(string(v43Name))
						out.RawByte(':')
						out.Bool(bool(v43Value))
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
			v44First := true
			for v44Name, v44Value := range in.NetworkLocs {
				if v44First {
					v44First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v44Name))
				out.RawByte(':')
				if v44Value == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
					out.RawString(`null`)
				} else {
					out.RawByte('{')
					v45First := true
					for v45Name, v45Value := range v44Value {
						if v45First {
							v45First = false
						} else {
							out.RawByte(',')
						}
						out.String(string(v45Name))
						out.RawByte(':')
						out.Bool(bool(v45Value))
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
			v46First := true
			for v46Name, v46Value := range in.MacVariability {
				if v46First {
					v46First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v46Name))
				out.RawByte(':')
				out.Float32(float32(v46Value))
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
			v47First := true
			for v47Name, v47Value := range in.MacCount {
				if v47First {
					v47First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v47Name))
				out.RawByte(':')
				out.Int(int(v47Value))
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
			v48First := true
			for v48Name, v48Value := range in.MacCountByLoc {
				if v48First {
					v48First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v48Name))
				out.RawByte(':')
				if v48Value == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
					out.RawString(`null`)
				} else {
					out.RawByte('{')
					v49First := true
					for v49Name, v49Value := range v48Value {
						if v49First {
							v49First = false
						} else {
							out.RawByte(',')
						}
						out.String(string(v49Name))
						out.RawByte(':')
						out.Int(int(v49Value))
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
			for v50, v51 := range in.UniqueLocs {
				if v50 > 0 {
					out.RawByte(',')
				}
				out.String(string(v51))
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
			for v52, v53 := range in.UniqueMacs {
				if v52 > 0 {
					out.RawByte(',')
				}
				out.String(string(v53))
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
			v54First := true
			for v54Name, v54Value := range in.LocCount {
				if v54First {
					v54First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v54Name))
				out.RawByte(':')
				out.Int(int(v54Value))
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
		case "GMutex":
			if in.IsNull() {
				in.Skip()
				out.GMutex = nil
			} else {
				if out.GMutex == nil {
					out.GMutex = new(sync.RWMutex)
				}
				easyjson3b8810b5DecodeSync(in, &*out.GMutex)
			}
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
		case "ConfigData":
			if in.IsNull() {
				in.Skip()
				out.ConfigData = nil
			} else {
				if out.ConfigData == nil {
					out.ConfigData = new(ConfigDataStruct)
				}
				(*out.ConfigData).UnmarshalEasyJSON(in)
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
		case "ConfigDataChanged":
			out.ConfigDataChanged = bool(in.Bool())
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
		const prefix string = ",\"GMutex\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.GMutex == nil {
			out.RawString("null")
		} else {
			easyjson3b8810b5EncodeSync(out, *in.GMutex)
		}
	}
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
		const prefix string = ",\"ConfigData\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.ConfigData == nil {
			out.RawString("null")
		} else {
			(*in.ConfigData).MarshalEasyJSON(out)
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
		const prefix string = ",\"ConfigDataChanged\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Bool(bool(in.ConfigDataChanged))
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
func easyjson3b8810b5DecodeSync(in *jlexer.Lexer, out *sync.RWMutex) {
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
func easyjson3b8810b5EncodeSync(out *jwriter.Writer, in sync.RWMutex) {
	out.RawByte('{')
	first := true
	_ = first
	out.RawByte('}')
}
func easyjson3b8810b5DecodeParsinServerDbm5(in *jlexer.Lexer, out *ConfigDataStruct) {
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
		case "KnnParameters":
			(out.KnnParameters).UnmarshalEasyJSON(in)
		case "GroupGraph":
			(out.GroupGraph).UnmarshalEasyJSON(in)
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
func easyjson3b8810b5EncodeParsinServerDbm5(out *jwriter.Writer, in ConfigDataStruct) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"KnnParameters\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		(in.KnnParameters).MarshalEasyJSON(out)
	}
	{
		const prefix string = ",\"GroupGraph\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		(in.GroupGraph).MarshalEasyJSON(out)
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ConfigDataStruct) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson3b8810b5EncodeParsinServerDbm5(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ConfigDataStruct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson3b8810b5EncodeParsinServerDbm5(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ConfigDataStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3b8810b5DecodeParsinServerDbm5(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ConfigDataStruct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson3b8810b5DecodeParsinServerDbm5(l, v)
}
func easyjson3b8810b5DecodeParsinServerDbm6(in *jlexer.Lexer, out *AlgoDataStruct) {
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
func easyjson3b8810b5EncodeParsinServerDbm6(out *jwriter.Writer, in AlgoDataStruct) {
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
	easyjson3b8810b5EncodeParsinServerDbm6(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v AlgoDataStruct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson3b8810b5EncodeParsinServerDbm6(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *AlgoDataStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3b8810b5DecodeParsinServerDbm6(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *AlgoDataStruct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson3b8810b5DecodeParsinServerDbm6(l, v)
}
