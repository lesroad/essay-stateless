package statistics

import (
	"essay-stateless/internal/model"
	"strconv"
	"strings"
)

func FromRequestModel(requests []model.StatisticsRequest) []StudentData {
	students := make([]StudentData, len(requests))

	for i, req := range requests {
		allScore, allTotal := parseWithTotal(req.ScoreEvaluation.Scores.AllWithTotal, req.ScoreEvaluation.Scores.All)
		contentScore, contentTotal := parseWithTotal(req.ScoreEvaluation.Scores.ContentWithTotal, req.ScoreEvaluation.Scores.Content)
		expressionScore, expressionTotal := parseWithTotal(req.ScoreEvaluation.Scores.ExpressionWithTotal, req.ScoreEvaluation.Scores.Expression)
		structureScore, structureTotal := parseWithTotal(req.ScoreEvaluation.Scores.StructureWithTotal, req.ScoreEvaluation.Scores.Structure)
		developmentScore, developmentTotal := parseWithTotal(req.ScoreEvaluation.Scores.DevelopmentWithTotal, req.ScoreEvaluation.Scores.Development)

		students[i] = StudentData{
			WordSentenceEvaluation: convertWordSentenceEvaluation(req.WordSentenceEvaluation),
			EssayScore: ScoreData{
				All:         allScore,
				Content:     contentScore,
				Expression:  expressionScore,
				Structure:   structureScore,
				Development: developmentScore,
			},
			EssayTotalScore: ScoreData{
				All:         allTotal,
				Content:     contentTotal,
				Expression:  expressionTotal,
				Structure:   structureTotal,
				Development: developmentTotal,
			},
		}
	}

	return students
}

// parseWithTotal 解析形如 "a/b" 的分数字符串，返回 (a, b)。
func parseWithTotal(withTotal string, fallbackScore int64) (score int, total int) {
	s := strings.TrimSpace(withTotal)
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return int(fallbackScore), 0
	}

	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])
	if left == "" || right == "" {
		return int(fallbackScore), 0
	}

	num, err1 := strconv.ParseInt(left, 10, 64)
	den, err2 := strconv.ParseInt(right, 10, 64)
	if err1 != nil || err2 != nil || den < 0 {
		return int(fallbackScore), 0
	}
	return int(num), int(den)
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

func ToResponseModel(result *AnalyzeResult, generatedTime int64) *model.ClassStatisticsResponse {
	return &model.ClassStatisticsResponse{
		SubmissionPercentage: result.SubmissionPercentage,
		OverallPerformance: model.OverallPerformance{
			AverageScore:         result.OverallPerformance.AverageScore,
			GradeDistribution:    convertGradeDistribution(result.OverallPerformance.GradeDistribution),
			SkillMasteryAnalysis: convertSkillMastery(result.OverallPerformance.SkillMasteryAnalysis),
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

type AnalyzeResult struct {
	SubmissionPercentage float64
	OverallPerformance   OverallPerformance
	ErrorAnalysis        ErrorAnalysis
	HighlightAnalysis    HighlightAnalysis
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
