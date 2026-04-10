package common

import "encoding/json"

type JSONObject map[string]json.RawMessage
type JSONArray []json.RawMessage

func DecodeObject(raw []byte) (JSONObject, error) {
	var out JSONObject
	err := json.Unmarshal(raw, &out)
	return out, err
}

func DecodeObjectRaw(raw json.RawMessage) (JSONObject, bool) {
	var out JSONObject
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false
	}
	return out, true
}

func DecodeArrayRaw(raw json.RawMessage) (JSONArray, bool) {
	var out JSONArray
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false
	}
	return out, true
}

func StringField(obj JSONObject, key string) string {
	raw, ok := obj[key]
	if !ok {
		return ""
	}
	var out string
	if err := json.Unmarshal(raw, &out); err != nil {
		return ""
	}
	return out
}

func FloatField(obj JSONObject, key string) (float64, bool) {
	raw, ok := obj[key]
	if !ok {
		return 0, false
	}
	var out float64
	if err := json.Unmarshal(raw, &out); err != nil {
		return 0, false
	}
	return out, true
}

func HasField(obj JSONObject, key string) bool {
	_, ok := obj[key]
	return ok
}
