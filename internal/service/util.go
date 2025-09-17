package service

import "math"

// 24（70/3≈23.333，向上取整为24）
// 20（60/3=20.0，向上取整为20）
// 3（5/2=2.5，向上取整为3）
// 2（4/2=2.0，向上取整为2）
func DivideAndRoundUp(numerator, denominator int64) int64 {
	return int64(math.Ceil(float64(numerator) / float64(denominator)))
}

// 23（70/3≈23.333，向下取整为23）
// 20（60/3=20.0，向下取整为20）
// 2（5/2=2.5，向下取整为2）
// 2（4/2=2.0，向下取整为2）
func DivideAndRoundDown(numerator, denominator int64) int64 {
	return int64(math.Floor(float64(numerator) / float64(denominator)))
}

