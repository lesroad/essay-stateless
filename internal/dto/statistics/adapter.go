package statistics

import "essay-stateless/internal/model"

// FromRequestModel 将请求模型转换为内部数据结构
func FromRequestModel(requests []model.StatisticsRequest) []StudentData {
	students := make([]StudentData, len(requests))

	for i, req := range requests {
		students[i] = StudentData{
			WordSentenceEvaluation: convertWordSentenceEvaluation(req.WordSentenceEvaluation),
			Scores: ScoreData{
				All:         int(req.ScoreEvaluation.Scores.All),
				Content:     int(req.ScoreEvaluation.Scores.Content),
				Expression:  int(req.ScoreEvaluation.Scores.Expression),
				Structure:   int(req.ScoreEvaluation.Scores.Structure),
				Development: int(req.ScoreEvaluation.Scores.Development),
			},
		}
	}

	return students
}

func convertWordSentenceEvaluation(src model.WordSentenceEvaluation) WordSentenceEvaluation {
	dst := WordSentenceEvaluation{
		SentenceEvaluations: make([][]SentenceEvaluation, len(src.SentenceEvaluations)),
	}

	for i, paragraphs := range src.SentenceEvaluations {
		dst.SentenceEvaluations[i] = make([]SentenceEvaluation, len(paragraphs))
		for j, sentence := range paragraphs {
			dst.SentenceEvaluations[i][j] = SentenceEvaluation{
				IsGoodSentence:  sentence.IsGoodSentence,
				Label:           sentence.Label,
				Type:            sentence.Type,
				WordEvaluations: convertWordEvaluations(sentence.WordEvaluations),
			}
		}
	}

	return dst
}

func convertWordEvaluations(src []model.WordEvaluation) []WordEvaluation {
	dst := make([]WordEvaluation, len(src))
	for i, word := range src {
		dst[i] = WordEvaluation{
			Span:    word.Span,
			Type:    word.Type,
			Ori:     word.Ori,
			Revised: word.Revised,
		}
	}
	return dst
}

// ToResponseModel 将分析结果转换为响应模型
func ToResponseModel(result *AnalyzeResult, generatedTime int64) *model.ClassStatisticsResponse {
	return &model.ClassStatisticsResponse{
		TotalStudents: result.TotalStudents,
		OverallPerformance: model.OverallPerformance{
			AverageScore:         result.OverallPerformance.AverageScore,
			GradeDistribution:    convertGradeDistribution(result.OverallPerformance.GradeDistribution),
			SkillMasteryAnalysis: convertSkillMastery(result.OverallPerformance.SkillMasteryAnalysis),
			Summary:              result.OverallPerformance.Summary,
		},
		ErrorAnalysis: model.ErrorAnalysis{
			ErrorDistribution: convertErrorDistribution(result.ErrorAnalysis.ErrorDistribution),
			ErrorTypeRatio:    convertErrorTypeRatio(result.ErrorAnalysis.ErrorTypeRatio),
			HighFrequencyList: convertHighFrequencyErrors(result.ErrorAnalysis.HighFrequencyList),
		},
		HighlightAnalysis: model.HighlightAnalysis{
			HighlightDistribution: convertHighlightDistribution(result.HighlightAnalysis.HighlightDistribution),
			HighlightTypeRatio:    convertHighlightTypeRatio(result.HighlightAnalysis.HighlightTypeRatio),
		},
		GeneratedTime: generatedTime,
	}
}

// AnalyzeResult 是从 domain/statistics 导入的
type AnalyzeResult struct {
	TotalStudents      int
	OverallPerformance OverallPerformance
	ErrorAnalysis      ErrorAnalysis
	HighlightAnalysis  HighlightAnalysis
}

func convertGradeDistribution(src []GradeDistributionItem) []model.GradeDistributionItem {
	dst := make([]model.GradeDistributionItem, len(src))
	for i, item := range src {
		dst[i] = model.GradeDistributionItem{
			Grade:        item.Grade,
			StudentCount: item.StudentCount,
			Percentage:   item.Percentage,
		}
	}
	return dst
}

func convertSkillMastery(src []SkillMasteryItem) []model.SkillMasteryItem {
	dst := make([]model.SkillMasteryItem, len(src))
	for i, item := range src {
		dst[i] = model.SkillMasteryItem{
			SkillName:         item.SkillName,
			GradeDistribution: convertSkillGradeDistribution(item.GradeDistribution),
		}
	}
	return dst
}

func convertSkillGradeDistribution(src []SkillGradeDistribution) []model.SkillGradeDistribution {
	dst := make([]model.SkillGradeDistribution, len(src))
	for i, item := range src {
		dst[i] = model.SkillGradeDistribution{
			Grade:        item.Grade,
			StudentCount: item.StudentCount,
			Percentage:   item.Percentage,
		}
	}
	return dst
}

func convertErrorDistribution(src []ErrorDistributionItem) []model.ErrorDistributionItem {
	dst := make([]model.ErrorDistributionItem, len(src))
	for i, item := range src {
		dst[i] = model.ErrorDistributionItem{
			ErrorCount:   item.ErrorCount,
			StudentCount: item.StudentCount,
			Percentage:   item.Percentage,
		}
	}
	return dst
}

func convertErrorTypeRatio(src []ErrorTypeItem) []model.ErrorTypeItem {
	dst := make([]model.ErrorTypeItem, len(src))
	for i, item := range src {
		dst[i] = model.ErrorTypeItem{
			ErrorType:    item.ErrorType,
			Count:        item.Count,
			Percentage:   item.Percentage,
			StudentCount: item.StudentCount,
		}
	}
	return dst
}

func convertHighFrequencyErrors(src []HighFrequencyError) []model.HighFrequencyError {
	dst := make([]model.HighFrequencyError, len(src))
	for i, item := range src {
		dst[i] = model.HighFrequencyError{
			ErrorText: item.ErrorText,
			ErrorType: item.ErrorType,
			Count:     item.Count,
			Examples:  item.Examples,
		}
	}
	return dst
}

func convertHighlightDistribution(src []HighlightDistributionItem) []model.HighlightDistributionItem {
	dst := make([]model.HighlightDistributionItem, len(src))
	for i, item := range src {
		dst[i] = model.HighlightDistributionItem{
			HighlightCount: item.HighlightCount,
			StudentCount:   item.StudentCount,
			Percentage:     item.Percentage,
		}
	}
	return dst
}

func convertHighlightTypeRatio(src []HighlightTypeItem) []model.HighlightTypeItem {
	dst := make([]model.HighlightTypeItem, len(src))
	for i, item := range src {
		dst[i] = model.HighlightTypeItem{
			HighlightType: item.HighlightType,
			Count:         item.Count,
			Percentage:    item.Percentage,
			StudentCount:  item.StudentCount,
		}
	}
	return dst
}
