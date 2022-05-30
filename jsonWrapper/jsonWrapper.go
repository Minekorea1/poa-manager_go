package jsonWrapper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type JsonWrapper struct {
	Data     map[string]interface{}
	strData  string
	byteData []byte
}

func NewJsonWrapper() *JsonWrapper {
	json := JsonWrapper{}
	json.Data = make(map[string]interface{})
	return &json
}

func (j *JsonWrapper) Marshal() bool {
	doc, err := json.MarshalIndent(j.Data, "", "    ")
	if err == nil {
		j.strData = string(doc)
		j.byteData = doc
		return true
	}

	return false
}

func (j *JsonWrapper) MarshalValue(v interface{}) bool {
	doc, err := json.MarshalIndent(v, "", "    ")
	if err == nil {
		j.Data = make(map[string]interface{})
		json.Unmarshal(doc, &j.Data)
		j.strData = string(doc)
		j.byteData = doc
		return true
	}

	return false
}

// func (j *JsonWrapper) ParseJsonText(jsonText string) bool {
// 	err := json.Unmarshal([]byte(jsonText), &j.Data)
// 	return err == nil
// }

func (j *JsonWrapper) ParseJson(jsonText string) bool {
	err := json.Unmarshal([]byte(jsonText), &j.Data)
	j.strData = jsonText
	j.byteData = []byte(jsonText)
	return err == nil
}

func (j *JsonWrapper) ParseJsonTo(jsonText string, v interface{}) bool {
	err := json.Unmarshal([]byte(jsonText), v)
	j.strData = jsonText
	j.byteData = []byte(jsonText)
	return err == nil
}

func (j *JsonWrapper) ToString() string {
	return j.strData
}

func (j *JsonWrapper) SetValue(key string, value interface{}) {
	j.Data[key] = value
}

func (j *JsonWrapper) GetRawValue(key string) interface{} {
	return jsonValue{j.Data[key]}
}

func (j *JsonWrapper) GetValue(key string) jsonValue {
	return jsonValue{j.Data[key]}
}

func (j *JsonWrapper) Print() {
	fmt.Println(j.strData)
}

func (j *JsonWrapper) ReadJson(path string) bool {
	jsonText, err := ioutil.ReadFile(path)
	if err == nil {
		j.ParseJsonTo(string(jsonText), &j.Data)
		fmt.Println(j.Data)
		return true
	}

	return false
}

func (j *JsonWrapper) ReadJsonTo(path string, v interface{}) bool {
	jsonText, err := ioutil.ReadFile(path)
	if err == nil {
		j.ParseJsonTo(string(jsonText), v)
		return true
	}

	return false
}

func (j *JsonWrapper) WriteJson(path string) {
	_ = ioutil.WriteFile(path, j.byteData, 0644)
}

type jsonValue struct {
	Value interface{}
}

func (v jsonValue) ToInt64() int64 {
	if v.Value == nil {
		return 0
	}
	return int64(v.Value.(float64))
}

func (v jsonValue) ToFloat64() float64 {
	if v.Value == nil {
		return 0
	}
	return v.Value.(float64)
}

func (v jsonValue) ToBool() bool {
	if v.Value == nil {
		return false
	}
	return v.Value.(bool)
}

func (v jsonValue) ToString() string {
	if v.Value == nil {
		return ""
	}
	return v.Value.(string)
}
