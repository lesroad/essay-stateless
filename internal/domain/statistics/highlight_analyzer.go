package statistics

import (
	"essay-stateless/internal/dto/statistics"
	"sort"
)

type HighlightAnalyzer struct {
	gradeCalc  *GradeCalculator
	typeMapper *HighlightTypeMapper
}

type HighlightTypeMapper struct{}

func (m *HighlightTypeMapper) MapType(label string) string {
	mappings := map[string]string{
		"排比": "排比",
		"比喻": "比喻",
		"拟人": "拟人",
		"夸张": "夸张",
		"立意": "立意",
		"好词": "语言",
		"好句": "语言",
		"语言": "语言",
		"思维": "思维",
		"修辞": "语言",
		"词汇": "语言",
		"句式": "语言",
	}

	if mapped, ok := mappings[label]; ok {
		return mapped
	}
	return "其他"
}

func NewHighlightAnalyzer(gradeCalc *GradeCalculator) *HighlightAnalyzer {
	return &HighlightAnalyzer{
		gradeCalc:  gradeCalc,
		typeMapper: &HighlightTypeMapper{},
	}
}

func (a *HighlightAnalyzer) Analyze(students []statistics.StudentData) statistics.HighlightAnalysis {
	studentHighlightCounts := make([]int, len(students))
	highlightTypeCount := make(map[string]int)
	highlightTypeStudents := make(map[string]map[int]bool)

	for idx, student := range students {
		highlightCount := 0
		highlightTypes := make(map[string]bool)

		for _, paragraphs := range student.WordSentenceEvaluation.SentenceEvaluations {
			for _, sentence := range paragraphs {
				// 好句分析
				if sentence.IsGoodSentence && sentence.Label != "" {
					highlightCount++
					highlightType := a.typeMapper.MapType(sentence.Label)

					highlightTypeCount[highlightType]++

					if !highlightTypes[highlightType] {
						highlightTypes[highlightType] = true
						if highlightTypeStudents[highlightType] == nil {
							highlightTypeStudents[highlightType] = make(map[int]bool)
						}
						highlightTypeStudents[highlightType][idx] = true
					}
				}

				// 好词分析
				for _, wordEval := range sentence.WordEvaluations {
					if level1, ok := wordEval.Type["level1"]; ok && level1 == "作文亮点" {
						if level2, ok := wordEval.Type["level2"]; ok {
							highlightCount++
							highlightType := a.typeMapper.MapType(level2)

							highlightTypeCount[highlightType]++

							if !highlightTypes[highlightType] {
								highlightTypes[highlightType] = true
								if highlightTypeStudents[highlightType] == nil {
									highlightTypeStudents[highlightType] = make(map[int]bool)
								}
								highlightTypeStudents[highlightType][idx] = true
							}
						}
					}
				}
			}
		}

		studentHighlightCounts[idx] = highlightCount
	}

	return statistics.HighlightAnalysis{
		HighlightDistribution: a.generateHighlightDistribution(studentHighlightCounts, len(students)),
		HighlightTypeRatio:    a.generateHighlightTypeRatio(highlightTypeCount, highlightTypeStudents, len(students)),
	}
}

func (a *HighlightAnalyzer) generateHighlightDistribution(counts []int, total int) []statistics.HighlightDistributionItem {
	distribution := make([]int, 7)

	for _, count := range counts {
		if count >= 6 {
			distribution[6]++
		} else {
			distribution[count]++
		}
	}

	labels := []string{"0个", "1个", "2个", "3个", "4个", "5个", "6个及以上"}
	result := make([]statistics.HighlightDistributionItem, 0, len(labels))

	for i, count := range distribution {
		result = append(result, statistics.HighlightDistributionItem{
			HighlightCount: labels[i],
			StudentCount:   count,
			Percentage:     a.gradeCalc.RoundPercentage(float64(count) / float64(total) * 100),
		})
	}

	return result
}

func (a *HighlightAnalyzer) generateHighlightTypeRatio(typeCount map[string]int, typeStudents map[string]map[int]bool, total int) []statistics.HighlightTypeItem {
	totalCount := 0
	for _, count := range typeCount {
		totalCount += count
	}

	allTypes := []string{"立意", "语言", "思维", "比喻", "拟人", "排比", "夸张", "其他"}
	result := make([]statistics.HighlightTypeItem, 0, len(allTypes))

	for _, highlightType := range allTypes {
		count := typeCount[highlightType]
		percentage := 0.0
		if totalCount > 0 {
			percentage = a.gradeCalc.RoundPercentage(float64(count) / float64(totalCount) * 100)
		}

		result = append(result, statistics.HighlightTypeItem{
			HighlightType: highlightType,
			Count:         count,
			Percentage:    percentage,
			StudentCount:  len(typeStudents[highlightType]),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	return result
}

