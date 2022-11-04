package gen

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/golang/glog"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Options are the options to set for rendering the template.
type Options struct {
	EnumsAsInts        bool
	EmitDefaults       bool
	OrigName           bool
	AllowUnknownFields bool
}

// TODO: Some field kinds require different libraries (ie "encoding/json" to escape strings),
// or "strconv" to convert ints to strings), but if a library is important and not used we'll
// get yelled at. Need to dynamically determine which libraries are required depending on the
// protobuf being templated.

// This function is called with a param which contains the entire definition of a method.
func ApplyTemplate(w io.Writer, f *protogen.File, opts Options) error {

	if err := headerTemplate.Execute(w, tplHeader{
		File:    f,
		Options: opts,
	}); err != nil {
		return err
	}

	return applyMessages(w, f.Messages, opts)
}

func applyMessages(w io.Writer, msgs []*protogen.Message, opts Options) error {
	for _, m := range msgs {

		if m.Desc.IsMapEntry() {
			glog.V(2).Infof("Skipping %s, mapentry message", m.GoIdent.GoName)
			continue
		}

		glog.V(2).Infof("Processing %s", m.GoIdent.GoName)
		m := tplMessage{
			Message: m,
			Options: opts,
		}
		if err := unmarshalTemplate.Execute(w, m); err != nil {
			return err
		}
		marshalTemplate := template.Must(template.New("marshal").Parse(
			genMarshalTemplate(m),
		))
		if err := marshalTemplate.Execute(w, m); err != nil {
			return err
		}

		if err := applyMessages(w, m.Messages, opts); err != nil {
			return err
		}

	}

	return nil
}

func genMarshalTemplate(msg tplMessage) string {
	funcStart := `
	// MarshalJSON implements json.Marshaler
func (msg *{{.GoIdent.GoName}}) MarshalJSON() ([]byte,error) {
	var buf bytes.Buffer
	var err error
	buf.WriteString("{")
`
	// Assumes there is a newline before
	funcEnd := `buf.WriteString("}")
	return buf.Bytes(), err
}`
	var sb strings.Builder
	titler := cases.Title(language.AmericanEnglish)

	// Handle different types of fields we may encounter
	// General guidelines:
	// 1. Prefer strconv functions over string formatting ones
	// 2. Use a string builder to join strings
	// 3. Don't worry about json indent formatting, since our main use case is hubble exporter
	//    which is consumed by some other logging/storage service

	// Helper functions
	addComma := func() {
		sb.WriteString(`buf.WriteString(",")` + "\n")
	}
	wrapString := func(wrapped string) string {
		return fmt.Sprintf(
			`buf.WriteString("\"")`+"\n\t%s\n\t"+`buf.WriteString("\"")`,
			wrapped,
		)
	}
	useMarshal := func(toMarshal string) string {
		return fmt.Sprintf(
			"res, err := json.Marshal(%s)\n\tif err != nil { return buf.Bytes(), err }\n\tbuf.Write(res)",
			toMarshal,
		)
	}
	getKindComment := func(kind protoreflect.Kind) string {
		return fmt.Sprintf("// %s", kind.GoString())
	}

	// Each string in this slice represents the line(s) of go code to write in order to
	// add a field into the json map.
	// This format makes it easier to put commas in the correct locations.
	templatedFields := []string{}

	// handleDecodeField handles anything that isn't "special"
	handleDecodeField := func(f *protogen.Field) {
		fieldKind := f.Desc.Kind()
		fieldJsonName := f.Desc.JSONName()
		fieldGoTitleName := titler.String(f.GoName)
		fieldGoGetter := fmt.Sprintf("msg.Get%s()", fieldGoTitleName)
		fieldValue := `buf.WriteString("\"\"")`
		switch fieldKind {
		case protoreflect.BoolKind:
			fieldValue = fmt.Sprintf("buf.WriteString(strconv.FormatBool(%s))", fieldGoGetter)
		case protoreflect.StringKind:
			fieldValue = useMarshal(fieldGoGetter)
		case protoreflect.BytesKind:
			fieldValue = wrapString(fmt.Sprintf("buf.WriteString(base64.StdEncoding.EncodeToString(%s))", fieldGoGetter))
		case protoreflect.Int32Kind:
			fieldValue = fmt.Sprintf("buf.WriteString(strconv.Itoa(int(%s)))", fieldGoGetter)
		case protoreflect.Int64Kind:
			fieldValue = wrapString(fmt.Sprintf("buf.WriteString(strconv.Itoa(int(%s)))", fieldGoGetter))
		case protoreflect.DoubleKind, protoreflect.FloatKind:
			fieldValue = useMarshal(fieldGoGetter)
		}
		templatedFields = append(templatedFields, fmt.Sprintf(
			"\t%s\n\t"+`buf.WriteString("\"%s\":")`+"\n\t%s\n\t",
			getKindComment(fieldKind), fieldJsonName, fieldValue,
		))
	}

	handleDecodeOneof := func(f *protogen.Oneof) {
	}

	handleDecodeEnum := func(f *protogen.Enum) {
	}

	handleDecodeNestedMessage := func(f *protogen.Message) {
	}

	handleDecodeExtension := func(f *protogen.Extension) {
	}

	// Message.Fields will contain every numbered field in a message, which makes it more difficult to just
	// use a single for loop over all fields.
	// For instance, if there is an Oneof in the message, than all possible values in the Oneof will be added
	// to the Fields slice. However, only a single struct representing the Oneof will be added to Message.Oneofs.
	// Breaking out field types into different for loops like this should help to reduce complexity, but it may
	// not be as pretty as a single for loop.
	for _, field := range msg.Message.Fields {
		if field.Oneof != nil || field.Enum != nil || field.Message != nil || field.Extendee != nil {
			continue
		}
		handleDecodeField(field)
	}
	for _, oneof := range msg.Message.Oneofs {
		handleDecodeOneof(oneof)
	}
	for _, enum := range msg.Message.Enums {
		handleDecodeEnum(enum)
	}
	for _, nestedMsg := range msg.Message.Messages {
		handleDecodeNestedMessage(nestedMsg)
	}
	for _, ext := range msg.Message.Extensions {
		handleDecodeExtension(ext)
	}

	// Write the template out
	sb.WriteString(funcStart)
	numLines := len(templatedFields)
	for i, line := range templatedFields {
		sb.WriteString(line)
		if i != numLines-1 {
			addComma()
		}
	}
	sb.WriteString(funcEnd)
	return sb.String()
}

type tplHeader struct {
	*protogen.File
	Options
}

type tplMessage struct {
	*protogen.Message
	Options
}

var (
	headerTemplate = template.Must(template.New("header").Parse(`
// Code generated by protoc-gen-go-json. DO NOT EDIT.
// source: {{.Proto.Name}}

package {{.GoPackageName}}

import (
	"bytes"
	"strconv"
	"encoding/base64"
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
)
`))

	unmarshalTemplate = template.Must(template.New("unmarshal").Parse(`
// UnmarshalJSON implements json.Unmarshaler
func (msg *{{.GoIdent.GoName}}) UnmarshalJSON(b []byte) error {
	return protojson.UnmarshalOptions {
		DiscardUnknown: {{.AllowUnknownFields}},
	}.Unmarshal(b, msg)
}
`))
)
