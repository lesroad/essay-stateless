package model

import (
	"encoding/json"
)

// EvaluateRequest 作文批改请求
type EvaluateRequest struct {
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	Grade     *int    `json:"grade,omitempty"`
	EssayType *string `json:"essayType,omitempty"`
}

func (r *EvaluateRequest) JSONString() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// OcrEvaluateRequest OCR作文批改请求
type OcrEvaluateRequest struct {
	Images    []string `json:"images"`
	LeftType  string   `json:"leftType"`
	Provider  *string  `json:"provider,omitempty"`
	ImageType *string  `json:"imageType,omitempty"`
	Grade     *int     `json:"grade,omitempty"`
	EssayType *string  `json:"essayType,omitempty"`
}

func (r *OcrEvaluateRequest) JSONString() string {
	data, _ := json.Marshal(r)
	return string(data)
}

type TitleOcrRequest struct {
	Images   []string `json:"images"`
	LeftType *string  `json:"leftType,omitempty"`
}

func (r *TitleOcrRequest) JSONString() string {
	data, err := json.Marshal(r)
	if err != nil {
		return "序列化失败"
	}
	return string(data)
}

type DefaultOcrRequest struct {
	Images   []string `json:"images"`
	LeftType *string  `json:"leftType,omitempty"`
}

func (r *DefaultOcrRequest) JSONString() string {
	data, err := json.Marshal(r)
	if err != nil {
		return "序列化失败"
	}
	return string(data)
}
