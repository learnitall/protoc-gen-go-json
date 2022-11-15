package main

import (
	"fmt"

	"github.com/golang/glog"
)

type Renderable interface {
	Render(*TemplateContext) ([]string, error)
}

type JsonRawBytes string

func (jrb JsonRawBytes) Render(ctx *TemplateContext) ([]string, error) {
	lines := []string{
		fmt.Sprintf(
			"%s.Write(%s)",
			ctx.BufferName, jrb,
		),
	}
	glog.V(1).Infof(
		"Rendered JsonRawBytes, in: %s, out: %s", jrb, lines,
	)
	return lines, nil
}

type JsonRaw string

func (jr JsonRaw) Render(ctx *TemplateContext) ([]string, error) {
	lines := []string{
		fmt.Sprintf(
			"%s.WriteString(%s)",
			ctx.BufferName, jr,
		),
	}
	glog.V(1).Infof(
		"Rendered JsonRaw, in: %s, out: %s", jr, lines,
	)
	return lines, nil
}

type JsonInt string

func (ji JsonInt) Render(ctx *TemplateContext) ([]string, error) {
	glog.V(1).Info("Adding import 'strconv' for JsonInt")
	ctx.AddImport("strconv")
	lines, err := JsonRaw("strconv.Itoa(int(" + ji + ")").Render(ctx)
	if err != nil {
		return nil, err
	}
	glog.V(1).Infof(
		"Rendered JsonInt, in: %s, out: %s", ji, lines,
	)
	return lines, err
}

type JsonFloat string

func (jf JsonFloat) Render(ctx *TemplateContext) ([]string, error) {
	glog.V(1).Info("Adding import 'strconv' for JsonFloat")
	ctx.AddImport("strconv")
	lines := []string{
		fmt.Sprintf(
			"res, err := strconv.ParseFloat(%s, 64)", jf,
		),
		"if err != nil {",
		"\treturn " + ctx.BufferName + ".Bytes(), err",
		"}",
		ctx.BufferName + ".WriteString(res)",
	}
	glog.V(1).Infof(
		"Rendered JsonFloat, in: %s, out: %s", jf, lines,
	)
	return lines, nil
}

type JsonBool string

func (jb JsonBool) Render(ctx *TemplateContext) ([]string, error) {
	glog.V(1).Info("Adding import 'strconv' for JsonBool")
	ctx.AddImport("strconv")
	lines, err := JsonRaw("strconv.FormatBool(" + jb + ")").Render(ctx)
	if err != nil {
		return nil, err
	}
	glog.V(1).Infof(
		"Rendered JsonBool, in: %s, out: %s", jb, lines,
	)
	return lines, nil
}

type JsonString string

func (js JsonString) Render(ctx *TemplateContext) ([]string, error) {
	lines, err := JsonRaw(`"` + js + `"`).Render(ctx)
	if err != nil {
		return nil, err
	}
	glog.V(1).Infof(
		"Rendered JsonString, in: %s, out: %s", js, lines,
	)
	return lines, nil
}

type JsonNull struct{}

func (jn JsonNull) Render(ctx *TemplateContext) ([]string, error) {
	lines, err := JsonRaw("null").Render(ctx)
	if err != nil {
		return nil, err
	}
	glog.V(1).Infof(
		"Rendered JsonNull, in: %s, out: %s", jn, lines,
	)
	return lines, nil
}

type JsonMarshal string

func (jm JsonMarshal) Render(ctx *TemplateContext) ([]string, error) {
	glog.V(1).Info("Adding import 'encoding/json' for JsonMarshal")
	ctx.AddImport("encoding/json")
	lines := []string{
		fmt.Sprintf(
			"res, err := json.Marshal(\"%s\")", jm,
		),
		"if err != nil {",
		"\treturn " + ctx.BufferName + ".Bytes(), err",
		"}",
		ctx.BufferName + ".WriteString(res)",
	}
	glog.V(1).Infof(
		"Rendered JsonMarshal, in: %s, out: %s", jm, lines,
	)
	return lines, nil
}

type JsonArray struct {
	Items []Renderable
}

func (ja *JsonArray) Render(ctx *TemplateContext) ([]string, error) {
	// open bracket lines
	openBracket, err := JsonString("[").Render(ctx)
	if err != nil {
		return nil, err
	}
	// close bracket lines
	closeBracket, err := JsonString("]").Render(ctx)
	if err != nil {
		return nil, err
	}
	// comma line
	comma, err := JsonString(",").Render(ctx)
	if err != nil {
		return nil, err
	}
	lines := []string{}
	numItems := len(ja.Items)

	lines = append(lines, openBracket...)
	for n, item := range ja.Items {
		itemLines, err := item.Render(ctx)
		if err != nil {
			return nil, err
		}
		lines = append(lines, itemLines...)
		if n != numItems-1 {
			lines = append(lines, comma...)
		}
	}
	lines = append(lines, closeBracket...)
	glog.V(1).Infof(
		"Rendered JsonArray, in: %s, out: %s", ja, lines,
	)
	return lines, nil
}

type JsonMapKeyPair struct {
	Key   Renderable
	Value Renderable
}

type JsonMap struct {
	KVPairs []JsonMapKeyPair
}

func (jm *JsonMap) Render(ctx *TemplateContext) ([]string, error) {
	openBrace, err := JsonString("{").Render(ctx)
	if err != nil {
		return nil, err
	}
	closeBrace, err := JsonString("}").Render(ctx)
	if err != nil {
		return nil, err
	}
	colon, err := JsonString(":").Render(ctx)
	if err != nil {
		return nil, err
	}
	comma, err := JsonString(",").Render(ctx)
	if err != nil {
		return nil, err
	}
	numPairs := len(jm.KVPairs)

	lines := []string{}
	lines = append(lines, openBrace...)
	for i, kv := range jm.KVPairs {
		renderedKey, err := kv.Key.Render(ctx)
		if err != nil {
			return nil, err
		}
		renderedValue, err := kv.Value.Render(ctx)
		if err != nil {
			return nil, err
		}
		lines = append(lines, renderedKey...)
		lines = append(lines, colon...)
		lines = append(lines, renderedValue...)
		if i != numPairs-1 {
			lines = append(lines, comma...)
		}
	}
	lines = append(lines, closeBrace...)
	glog.V(1).Infof(
		"Rendered JsonMap, in: %s, out: %s", jm, lines,
	)
	return lines, nil
}
