package statistics

import (
	"essay-stateless/internal/dto/statistics"
	"math"
)

type OverallAnalyzer struct {
	gradeCalc *GradeCalculator
}

func NewOverallAnalyzer(gradeCalc *GradeCalculator) *OverallAnalyzer {
	return &OverallAnalyzer{
		gradeCalc: gradeCalc,
	}
}

func (a *OverallAnalyzer) Analyze(students []statistics.StudentData) statistics.OverallPerformance {
	totalStudents := len(students)

	// 计算平均分
	averageScore := a.calculateAverageScore(students)

	// 生成等级分布
	gradeDistribution := a.generateGradeDistribution(students, totalStudents)

	// 各个技能的等级分布
	skillMastery := a.generateSkillMastery(students)

	return statistics.OverallPerformance{
		AverageScore:         math.Round(averageScore*100) / 100,
		GradeDistribution:    gradeDistribution,
		SkillMasteryAnalysis: skillMastery,
	}
}

func (a *OverallAnalyzer) calculateAverageScore(students []statistics.StudentData) float64 {
	var total float64
	for _, student := range students {
		total += float64(student.EssayScore.All)
	}
	return total / float64(len(students))
}

func (a *OverallAnalyzer) generateGradeDistribution(students []statistics.StudentData, total int) []statistics.GradeDistributionItem {
	gradeCount := make(map[string]int)

	for _, student := range students {
		grade := a.gradeCalc.CalculateGrade(float64(student.EssayScore.All), float64(student.EssayTotalScore.All))
		gradeCount[grade]++
	}

	grades := []string{"优秀", "良好", "合格", "不合格"}
	result := make([]statistics.GradeDistributionItem, 0, len(grades))

	for _, grade := range grades {
		count := gradeCount[grade]
		percentage := a.gradeCalc.RoundPercentage(float64(count) / float64(total) * 100)

		result = append(result, statistics.GradeDistributionItem{
			Grade:        grade,
			StudentCount: count,
			Percentage:   percentage,
		})
	}

	return result
}

func (a *OverallAnalyzer) generateSkillMastery(students []statistics.StudentData) []statistics.SkillMasteryItem {
	skills := []struct {
		name     string
		maxScore float64
		getScore func(statistics.StudentData) float64
	}{
		{"All", float64(students[0].EssayTotalScore.All), func(s statistics.StudentData) float64 { return float64(s.EssayScore.All) }},
		{"Content", float64(students[0].EssayTotalScore.Content), func(s statistics.StudentData) float64 { return float64(s.EssayScore.Content) }},
		{"Expression", float64(students[0].EssayTotalScore.Expression), func(s statistics.StudentData) float64 { return float64(s.EssayScore.Expression) }},
		{"Structure", float64(students[0].EssayTotalScore.Structure), func(s statistics.StudentData) float64 { return float64(s.EssayScore.Structure) }},
		{"Development", float64(students[0].EssayTotalScore.Development), func(s statistics.StudentData) float64 { return float64(s.EssayScore.Development) }},
	}

	grades := []string{"优秀", "良好", "合格", "不合格"}
	result := make([]statistics.SkillMasteryItem, 0, len(skills))

	for _, skill := range skills {
		gradeCount := make(map[string]int)

		for _, student := range students {
			score := skill.getScore(student)
			grade := a.gradeCalc.CalculateGrade(score, skill.maxScore)
			gradeCount[grade]++
		}

		gradeDistribution := make([]statistics.SkillGradeDistribution, 0, len(grades))
		for _, grade := range grades {
			count := gradeCount[grade]
			percentage := 0.0

			gradeDistribution = append(gradeDistribution, statistics.SkillGradeDistribution{
				Grade:        grade,
				StudentCount: count,
				Percentage:   percentage,
			})
		}

		result = append(result, statistics.SkillMasteryItem{
			SkillName:         skill.name,
			GradeDistribution: gradeDistribution,
		})
	}

	return result
}
