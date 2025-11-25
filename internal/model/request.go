package model

import (
	"encoding/json"
)

// EvaluateRequest 作文批改请求
type EvaluateRequest struct {
	Title            string  `json:"title"`
	Content          string  `json:"content"`
	Grade            *int    `json:"grade,omitempty"`
	EssayType        *string `json:"essayType,omitempty"`
	TotalScore       *int64  `json:"totalScore,omitempty"`
	Prompt           *string `json:"prompt,omitempty"` // 题干/写作要求
	Standard         *string `json:"standard,omitempty"`
	ContentScore     *int64  `json:"contentScore,omitempty"`
	ExpressionScore  *int64  `json:"expressionScore,omitempty"`
	StructureScore   *int64  `json:"structureScore,omitempty"`
	DevelopmentScore *int64  `json:"developmentScore,omitempty"`
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

// StatisticsRequest 单个学生的统计分析请求
type StatisticsRequest struct {
	WordSentenceEvaluation WordSentenceEvaluation `json:"wordSentenceEvaluation,omitempty"` // 好词好句评价
	ScoreEvaluation        ScoreEvaluation        `json:"scoreEvaluations,omitempty"`       // 分数点评
}

// ClassStatisticsRequest 班级学情统计分析请求 - 直接传入学生数据数组
type ClassStatisticsRequest []StatisticsRequest

func (r ClassStatisticsRequest) JSONString() string {
	data, _ := json.Marshal(r)
	return string(data)
}
