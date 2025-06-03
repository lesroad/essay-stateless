package model

import (
	"encoding/json"
)

type BetaEvaluateRequest struct {
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	Grade     *int    `json:"grade,omitempty"`
	EssayType *string `json:"essayType,omitempty"`
}

func (r *BetaEvaluateRequest) JSONString() string {
	data, err := json.Marshal(r)
	if err != nil {
		return "序列化失败"
	}
	return string(data)
}

type BetaOcrEvaluateRequest struct {
	Images    []string `json:"images"`
	LeftType  *string  `json:"leftType,omitempty"`
	ImageType *string  `json:"imageType,omitempty"`
	Provider  *string  `json:"provider,omitempty"`
	Grade     *int     `json:"grade,omitempty"`
	EssayType *string  `json:"essayType,omitempty"`
}

func (r *BetaOcrEvaluateRequest) JSONString() string {
	data, err := json.Marshal(r)
	if err != nil {
		return "序列化失败"
	}
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
