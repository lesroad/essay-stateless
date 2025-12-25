package statistics

import "math"

// GradeCalculator 等级计算器
type GradeCalculator struct {
	config GradeConfig
}

// GradeConfig 等级划分配置
type GradeConfig struct {
	ExcellentMin float64 // 优秀最低分（百分比）
	GoodMin      float64 // 良好最低分（百分比）
	PassMin      float64 // 合格最低分（百分比）
}

// NewGradeCalculator 创建等级计算器
func NewGradeCalculator() *GradeCalculator {
	return &GradeCalculator{
		config: GradeConfig{
			ExcellentMin: 0.9, // 90%及以上为优秀
			GoodMin:      0.8, // 80%-89%为良好
			PassMin:      0.6, // 60%-79%为合格
		},
	}
}

func (c *GradeCalculator) CalculateGrade(score, maxScore float64) string {
	if maxScore == 0 || score == 0 {
		return "不合格"
	}

	percentage := score / maxScore

	if percentage >= c.config.ExcellentMin {
		return "优秀"
	} else if percentage >= c.config.GoodMin {
		return "良好"
	} else if percentage >= c.config.PassMin {
		return "合格"
	}
	return "不合格"
}

// RoundPercentage 四舍五入百分比
func (c *GradeCalculator) RoundPercentage(percentage float64) float64 {
	return math.Round(percentage*100) / 100
}

