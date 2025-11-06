package evaluate

import "unicode/utf8"

// GrammarPosition 语法位置
type GrammarPosition struct {
	ParagraphIndex int
	SentenceIndex  int
	RelativeIndex  int
}

// PositionCalculator 位置计算器
type PositionCalculator struct{}

// NewPositionCalculator 创建位置计算器
func NewPositionCalculator() *PositionCalculator {
	return &PositionCalculator{}
}

// GetSentenceRelativeIndex 获取字符在句子中的相对位置
//
// 根据全局字符偏移量，计算出该字符位于哪个段落的哪个句子，以及在句子中的相对位置
//
// 示例：
//
//	text = [
//	  ["第一段第一句", "第一段第二句"],
//	  ["第二段第一句", "第二段第二句"]
//	]
//
//	startPos = 8
//	第一段第一句：位置 0-4 (长度5)
//	第一段第二句：位置 5-9 (长度5)
//	startPos=8 落在第一段第二句中，相对位置是 8-5=3
//
//	返回: {ParagraphIndex: 0, SentenceIndex: 1, RelativeIndex: 3}
func (c *PositionCalculator) GetSentenceRelativeIndex(text [][]string, startPos int) *GrammarPosition {
	currentPos := 0
	for pIndex, paragraph := range text {
		for sIndex, sentence := range paragraph {
			sentenceLen := utf8.RuneCountInString(sentence)
			if startPos >= currentPos && startPos < currentPos+sentenceLen {
				return &GrammarPosition{
					ParagraphIndex: pIndex,
					SentenceIndex:  sIndex,
					RelativeIndex:  startPos - currentPos,
				}
			}
			currentPos += sentenceLen
		}
		currentPos += 1 // 段落间的换行符
	}
	return nil
}

