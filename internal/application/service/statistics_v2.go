package service

import (
	"context"
	"time"

	"essay-stateless/internal/domain/statistics"
	statistics_dto "essay-stateless/internal/dto/statistics"
	"essay-stateless/internal/model"
)

// StatisticsServiceV2 新版学情统计服务（基于DDD架构）
type StatisticsServiceV2 struct {
	analyzer *statistics.Analyzer
}

// NewStatisticsServiceV2 创建新版学情统计服务
func NewStatisticsServiceV2() *StatisticsServiceV2 {
	return &StatisticsServiceV2{
		analyzer: statistics.NewAnalyzer(),
	}
}

// AnalyzeClassStatistics 分析班级学情统计
func (s *StatisticsServiceV2) AnalyzeClassStatistics(ctx context.Context, req []model.StatisticsRequest) (*model.ClassStatisticsResponse, error) {
	if len(req) == 0 {
		return nil, ErrEmptyStudentData
	}

	// 转换为内部数据结构
	studentData := statistics_dto.FromRequestModel(req)

	// 执行领域分析
	result := s.analyzer.Analyze(studentData)

	// 转换为响应模型
	response := statistics_dto.ToResponseModel(&statistics_dto.AnalyzeResult{
		TotalStudents:      result.TotalStudents,
		OverallPerformance: result.OverallPerformance,
		ErrorAnalysis:      result.ErrorAnalysis,
		HighlightAnalysis:  result.HighlightAnalysis,
	}, time.Now().Unix())

	return response, nil
}

// ErrEmptyStudentData 空学生数据错误
var ErrEmptyStudentData = &ServiceError{
	Code:    "EMPTY_STUDENT_DATA",
	Message: "学生数据不能为空",
}

// ServiceError 服务层错误
type ServiceError struct {
	Code    string
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}
