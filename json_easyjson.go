// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package json

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

func easyjson42239ddeDecodeGithubComLibforJson(in *jlexer.Lexer, out *EasyType) {
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
		case "Food":
			out.Food = string(in.String())
		case "Tags":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.Tags = make(map[string]string)
				} else {
					out.Tags = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v1 string
					v1 = string(in.String())
					(out.Tags)[key] = v1
					in.WantComma()
				}
				in.Delim('}')
			}
		case "Nested":
			if in.IsNull() {
				in.Skip()
				out.Nested = nil
			} else {
				if out.Nested == nil {
					out.Nested = new(Nested)
				}
				easyjson42239ddeDecodeGithubComLibforJson1(in, &*out.Nested)
			}
		case "SomeList":
			if in.IsNull() {
				in.Skip()
				out.SomeList = nil
			} else {
				in.Delim('[')
				if out.SomeList == nil {
					if !in.IsDelim(']') {
						out.SomeList = make([]string, 0, 4)
					} else {
						out.SomeList = []string{}
					}
				} else {
					out.SomeList = (out.SomeList)[:0]
				}
				for !in.IsDelim(']') {
					var v2 string
					v2 = string(in.String())
					out.SomeList = append(out.SomeList, v2)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "EmptyList":
			if in.IsNull() {
				in.Skip()
				out.EmptyList = nil
			} else {
				in.Delim('[')
				if out.EmptyList == nil {
					if !in.IsDelim(']') {
						out.EmptyList = make([]string, 0, 4)
					} else {
						out.EmptyList = []string{}
					}
				} else {
					out.EmptyList = (out.EmptyList)[:0]
				}
				for !in.IsDelim(']') {
					var v3 string
					v3 = string(in.String())
					out.EmptyList = append(out.EmptyList, v3)
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
func easyjson42239ddeEncodeGithubComLibforJson(out *jwriter.Writer, in EasyType) {
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
		const prefix string = ",\"Food\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Food))
	}
	{
		const prefix string = ",\"Tags\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Tags == nil && (out.Flags&jwriter.NilMapAsEmpty) == 0 {
			out.RawString(`null`)
		} else {
			out.RawByte('{')
			v4First := true
			for v4Name, v4Value := range in.Tags {
				if v4First {
					v4First = false
				} else {
					out.RawByte(',')
				}
				out.String(string(v4Name))
				out.RawByte(':')
				out.String(string(v4Value))
			}
			out.RawByte('}')
		}
	}
	{
		const prefix string = ",\"Nested\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Nested == nil {
			out.RawString("null")
		} else {
			easyjson42239ddeEncodeGithubComLibforJson1(out, *in.Nested)
		}
	}
	{
		const prefix string = ",\"SomeList\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.SomeList == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v5, v6 := range in.SomeList {
				if v5 > 0 {
					out.RawByte(',')
				}
				out.String(string(v6))
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"EmptyList\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.EmptyList == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v7, v8 := range in.EmptyList {
				if v7 > 0 {
					out.RawByte(',')
				}
				out.String(string(v8))
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v EasyType) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson42239ddeEncodeGithubComLibforJson(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v EasyType) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson42239ddeEncodeGithubComLibforJson(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *EasyType) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson42239ddeDecodeGithubComLibforJson(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *EasyType) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson42239ddeDecodeGithubComLibforJson(l, v)
}
func easyjson42239ddeDecodeGithubComLibforJson1(in *jlexer.Lexer, out *Nested) {
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
		case "Amazing":
			out.Amazing = string(in.String())
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
func easyjson42239ddeEncodeGithubComLibforJson1(out *jwriter.Writer, in Nested) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"Amazing\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Amazing))
	}
	out.RawByte('}')
}