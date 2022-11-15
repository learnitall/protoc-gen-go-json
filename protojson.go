package main

import (
	"fmt"

	"github.com/golang/glog"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func getMsgFieldGetter(ctx *TemplateContext, field *protogen.Field) string {
	return fmt.Sprintf(
		"%s.Get%s()",
		ctx.MessageName,
		cases.Title(language.AmericanEnglish).String(field.GoName),
	)
}

type ProtoJsonMessage struct {
	JsonMap
}

type ProtoJsonEnum struct {
}

type ProtoJsonMap struct {
}

type ProtoJsonRepeated struct {
}

type ProtoJsonBool = JsonBool

func NewProtoJsonBool(ctx *TemplateContext, field *protogen.Field) *ProtoJsonBool {
	r := ProtoJsonBool(JsonBool(getMsgFieldGetter(ctx, field)))
	return &r
}

type ProtoJsonString = JsonRaw

func NewProtoJsonString(ctx *TemplateContext, field *protogen.Field) *ProtoJsonString {
	r := ProtoJsonString(JsonRaw(getMsgFieldGetter(ctx, field)))
	return &r
}

type ProtoJsonBytes = JsonRaw

func NewProtoJsonBytes(ctx *TemplateContext, field *protogen.Field) *ProtoJsonBytes {
	glog.V(1).Info("Adding import 'encoding/base64' for ProtoJsonBytes")
	ctx.AddImport("encoding/base64")
	r := ProtoJsonBytes(JsonRaw(
		fmt.Sprintf("base64.StdEncoding.EncodeToString(%s)", getMsgFieldGetter(ctx, field)),
	))
	return &r
}

type ProtoJsonInt32 struct {
}

type ProtoJsonInt64 struct {
}

type ProtoJsonFloat struct {
}

type ProtoJsonAny struct {
}

type ProtoJsonTimestamp struct {
}

type ProtoJsonDuration struct {
}

type ProtoJsonStruct struct {
}

type ProtoJsonFieldMask struct {
}

type ProtoJsonListValue struct {
}

type ProtoJsonNullValue struct {
}

type ProtoJsonEmpty struct {
}

func ConstructMessageTree(ctx *TemplateContext, msg *protogen.Message) *ProtoJsonMessage {
	glog.V(1).Infof(
		"Creating json tree for %s from %s",
		ctx.RenderMessage.GoIdent.GoName,
		ctx.RenderMessage.Location.SourceFile,
	)
	jmsg := &ProtoJsonMessage{}
	for _, field := range msg.Fields {
		fieldJsonName := JsonString(`\"` + field.Desc.JSONName() + `\"`)
		var value Renderable
		switch field.Desc.Kind() {
		case protoreflect.BoolKind:
			value = NewProtoJsonBool(ctx, field)
		case protoreflect.StringKind:
			value = NewProtoJsonString(ctx, field)
		case protoreflect.BytesKind:
			value = NewProtoJsonBytes(ctx, field)
		}
		jmsg.KVPairs = append(jmsg.KVPairs, JsonMapKeyPair{Key: fieldJsonName, Value: value})
	}
	return jmsg
}
