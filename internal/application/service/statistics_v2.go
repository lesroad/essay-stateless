package service

import (
	"context"
	"time"

	"essay-stateless/internal/domain/statistics"
	statistics_dto "essay-stateless/internal/dto/statistics"
	"essay-stateless/internal/model"
)

type StatisticsServiceV2 struct {
	analyzer *statistics.Analyzer
}

func NewStatisticsServiceV2() *StatisticsServiceV2 {
	return &StatisticsServiceV2{
		analyzer: statistics.NewAnalyzer(),
	}
}

func (s *StatisticsServiceV2) AnalyzeClassStatistics(ctx context.Context, req model.ClassStatisticsRequest) (*model.ClassStatisticsResponse, error) {
	studentData := statistics_dto.FromRequestModel(req.SubmittedStudents)
	result := s.analyzer.Analyze(studentData, req.TotalStudents)
	response := statistics_dto.ToResponseModel(&statistics_dto.AnalyzeResult{
		SubmissionPercentage: result.SubmissionPercentage,
		OverallPerformance:   result.OverallPerformance,
		ErrorAnalysis:        result.ErrorAnalysis,
		HighlightAnalysis:    result.HighlightAnalysis,
	}, time.Now().Unix())

	return response, nil
}
