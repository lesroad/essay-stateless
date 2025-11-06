package evaluate

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// ScoreCalculator 分数计算器
type ScoreCalculator struct{}

// NewScoreCalculator 创建分数计算器
func NewScoreCalculator() *ScoreCalculator {
	return &ScoreCalculator{}
}

// CalculateAllScore 计算总分
func (c *ScoreCalculator) CalculateAllScore(score int64, maxScore, totalScore int64) (int64, string) {
	allScore := decimal.NewFromInt(score).
		Div(decimal.NewFromInt(maxScore)).
		Mul(decimal.NewFromInt(totalScore)).
		Round(0).
		IntPart()

	return allScore, fmt.Sprintf("%d/%d", allScore, totalScore)
}

// CalculateSubScore 计算子项分数（内容、表达等）
func (c *ScoreCalculator) CalculateSubScore(score int64, maxScore, totalScore int64, roundType string) (int64, string) {
	var subTotalScore int64

	switch roundType {
	case "up":
		subTotalScore = c.divideAndRoundUp(totalScore, 3)
	case "down":
		subTotalScore = c.divideAndRoundDown(totalScore, 3)
	default:
		subTotalScore = totalScore / 3
	}

	subScore := decimal.NewFromInt(score).
		Div(decimal.NewFromInt(maxScore)).
		Mul(decimal.NewFromInt(subTotalScore)).
		Round(0).
		IntPart()

	return subScore, fmt.Sprintf("%d/%d", subScore, subTotalScore)
}

// CalculateRemainderScore 计算剩余分数（用于结构或发展）
func (c *ScoreCalculator) CalculateRemainderScore(totalScore, contentScore, expressionScore int64, totalMax int64) (int64, string) {
	subTotalScore := c.divideAndRoundDown(totalMax, 3)
	remainderScore := totalScore - contentScore - expressionScore

	return remainderScore, fmt.Sprintf("%d/%d", remainderScore, subTotalScore)
}

// divideAndRoundUp 除法向上取整
func (c *ScoreCalculator) divideAndRoundUp(a, b int64) int64 {
	return (a + b - 1) / b
}

// divideAndRoundDown 除法向下取整
func (c *ScoreCalculator) divideAndRoundDown(a, b int64) int64 {
	return a / b
}

