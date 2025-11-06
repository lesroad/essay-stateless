package statistics

import (
	"essay-stateless/internal/dto/statistics"
	"math"
)

// OverallAnalyzer 整体表现分析器
type OverallAnalyzer struct {
	gradeCalc *GradeCalculator
}

// NewOverallAnalyzer 创建整体表现分析器
func NewOverallAnalyzer(gradeCalc *GradeCalculator) *OverallAnalyzer {
	return &OverallAnalyzer{
		gradeCalc: gradeCalc,
	}
}

// Analyze 分析整体表现
func (a *OverallAnalyzer) Analyze(students []statistics.StudentData) statistics.OverallPerformance {
	totalStudents := len(students)
	if totalStudents == 0 {
		return statistics.OverallPerformance{}
	}

	// 计算平均分
	averageScore := a.calculateAverageScore(students)

	// 生成等级分布
	gradeDistribution := a.generateGradeDistribution(students, totalStudents)

	// 生成写作技能掌握情况
	skillMastery := a.generateSkillMastery(students, totalStudents)

	// 生成总结
	summary := a.generateSummary(averageScore, gradeDistribution, totalStudents)

	return statistics.OverallPerformance{
		AverageScore:         math.Round(averageScore*100) / 100,
		GradeDistribution:    gradeDistribution,
		SkillMasteryAnalysis: skillMastery,
		Summary:              summary,
	}
}

func (a *OverallAnalyzer) calculateAverageScore(students []statistics.StudentData) float64 {
	var total float64
	for _, student := range students {
		total += float64(student.Scores.All)
	}
	return total / float64(len(students))
}

func (a *OverallAnalyzer) generateGradeDistribution(students []statistics.StudentData, total int) []statistics.GradeDistributionItem {
	gradeCount := make(map[string]int)

	for _, student := range students {
		grade := a.gradeCalc.CalculateGrade(float64(student.Scores.All), 90.0)
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

func (a *OverallAnalyzer) generateSkillMastery(students []statistics.StudentData, total int) []statistics.SkillMasteryItem {
	skills := []struct {
		name     string
		maxScore float64
		getScore func(statistics.StudentData) float64
	}{
		{"All", 90.0, func(s statistics.StudentData) float64 { return float64(s.Scores.All) }},
		{"Appearance", 30.0, func(s statistics.StudentData) float64 { return float64(s.Scores.Structure) }},
		{"Content", 30.0, func(s statistics.StudentData) float64 { return float64(s.Scores.Content) }},
		{"Expression", 30.0, func(s statistics.StudentData) float64 { return float64(s.Scores.Expression) }},
		{"Structure", 30.0, func(s statistics.StudentData) float64 { return float64(s.Scores.Structure) }},
		{"Development", 30.0, func(s statistics.StudentData) float64 { return float64(s.Scores.Development) }},
	}

	grades := []string{"优秀", "良好", "合格", "不合格"}
	result := make([]statistics.SkillMasteryItem, 0, len(skills))

	for _, skill := range skills {
		gradeCount := make(map[string]int)
		validCount := 0

		// 统计各等级分布（排除0分）
		for _, student := range students {
			score := skill.getScore(student)
			if score == 0 {
				continue
			}
			validCount++
			grade := a.gradeCalc.CalculateGrade(score, skill.maxScore)
			gradeCount[grade]++
		}

		// 生成等级分布
		gradeDistribution := make([]statistics.SkillGradeDistribution, 0, len(grades))
		for _, grade := range grades {
			count := gradeCount[grade]
			percentage := 0.0
			if validCount > 0 {
				percentage = a.gradeCalc.RoundPercentage(float64(count) / float64(validCount) * 100)
			}

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

func (a *OverallAnalyzer) generateSummary(avgScore float64, distribution []statistics.GradeDistributionItem, total int) string {
	level := a.gradeCalc.CalculateGrade(avgScore, 90.0)

	var excellentRate float64
	var poorCount int
	for _, item := range distribution {
		if item.Grade == "优秀" {
			excellentRate = item.Percentage
		} else if item.Grade == "不合格" {
			poorCount = item.StudentCount
		}
	}

	return formatSummary(level, avgScore, excellentRate, poorCount)
}

func formatSummary(level string, avgScore, excellentRate float64, poorCount int) string {
	return "班级整体表现" + level + "，平均分" + formatFloat(avgScore) + "分。其中" +
		formatFloat(excellentRate) + "%的学生达到优秀水平，还有" +
		formatInt(poorCount) + "名学生需要重点关注和帮助。"
}

func formatFloat(v float64) string {
	return trimZeros(formatFloatWithPrecision(v, 1))
}

func formatInt(v int) string {
	s := ""
	n := v
	for {
		s = string(rune('0'+(n%10))) + s
		n /= 10
		if n == 0 {
			break
		}
	}
	return s
}

func formatFloatWithPrecision(v float64, precision int) string {
	// 简单的浮点数格式化
	intPart := int(v)
	fracPart := int((v - float64(intPart)) * math.Pow(10, float64(precision)))
	return formatInt(intPart) + "." + formatInt(fracPart)
}

func trimZeros(s string) string {
	// 去除小数点后多余的0
	if len(s) == 0 {
		return s
	}
	hasPoint := false
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			hasPoint = true
			break
		}
	}
	if !hasPoint {
		return s
	}
	// 从后往前去除0
	i := len(s) - 1
	for i >= 0 && s[i] == '0' {
		i--
	}
	if i >= 0 && s[i] == '.' {
		i--
	}
	return s[:i+1]
}

