package evaluate

import (
	dto_evaluate "essay-stateless/internal/dto/evaluate"
	"essay-stateless/internal/model"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jinzhu/copier"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

// ResponseProcessor 响应处理器
type ResponseProcessor struct {
	cleaner *ContentCleaner
	posCalc *PositionCalculator
}

// NewResponseProcessor 创建响应处理器
func NewResponseProcessor() *ResponseProcessor {
	return &ResponseProcessor{
		cleaner: NewContentCleaner(),
		posCalc: NewPositionCalculator(),
	}
}

// ProcessWordSentence 处理词句评估响应
func (p *ResponseProcessor) ProcessWordSentence(wordSentence *dto_evaluate.APIWordSentence, response *model.EvaluateResponse) {
	if wordSentence == nil {
		return
	}

	sentencesEvaluations := make([][]model.SentenceEvaluation, len(response.Text))
	for i, paragraph := range response.Text {
		sentencesEvaluations[i] = make([]model.SentenceEvaluation, len(paragraph))
		for j := range paragraph {
			sentencesEvaluations[i][j] = model.SentenceEvaluation{
				IsGoodSentence:  false,
				Label:           "",
				Type:            make(map[string]string),
				WordEvaluations: []model.WordEvaluation{},
			}
		}
	}

	// 处理好句
	for _, sent := range wordSentence.Data.Results.GoodSents {
		if sent.ParagraphID < len(sentencesEvaluations) && sent.SentID < len(sentencesEvaluations[sent.ParagraphID]) {
			sentenceEval := &sentencesEvaluations[sent.ParagraphID][sent.SentID]
			sentenceEval.IsGoodSentence = true
			sentenceEval.Label = sent.Label
			sentenceEval.Type["level1"] = "作文亮点"
			sentenceEval.Type["level2"] = "好句"
		}
	}

	// 处理好词
	for _, word := range wordSentence.Data.Results.GoodWords {
		if word.ParagraphID < len(sentencesEvaluations) && word.SentID < len(sentencesEvaluations[word.ParagraphID]) {
			wordEval := model.WordEvaluation{
				Span: []int{word.Start, word.End},
				Type: map[string]string{
					"level1": "作文亮点",
					"level2": "好词",
				},
			}
			sentencesEvaluations[word.ParagraphID][word.SentID].WordEvaluations = append(
				sentencesEvaluations[word.ParagraphID][word.SentID].WordEvaluations, wordEval)
		}
	}

	response.AIEvaluation.WordSentenceEvaluation.SentenceEvaluations = sentencesEvaluations
	response.AIEvaluation.WordSentenceEvaluation.WordSentenceScore = wordSentence.Score
}

// ProcessGrammar 处理语法检查响应
func (p *ResponseProcessor) ProcessGrammar(grammarInfo *dto_evaluate.APIGrammarInfo, response *model.EvaluateResponse) {
	if grammarInfo == nil {
		return
	}

	for _, typo := range grammarInfo.Grammar.Typo {
		gp := p.posCalc.GetSentenceRelativeIndex(response.Text, typo.StartPos)
		if gp == nil {
			continue
		}

		wordEval := model.WordEvaluation{
			Span: []int{gp.RelativeIndex, gp.RelativeIndex + typo.EndPos - typo.StartPos},
			Type: map[string]string{
				"level1": "还需努力",
				"level2": typo.Type,
			},
			Ori:     typo.Ori,
			Revised: typo.Revised,
		}

		response.AIEvaluation.WordSentenceEvaluation.SentenceEvaluations[gp.ParagraphIndex][gp.SentenceIndex].WordEvaluations = append(
			response.AIEvaluation.WordSentenceEvaluation.SentenceEvaluations[gp.ParagraphIndex][gp.SentenceIndex].WordEvaluations, wordEval)
	}
}

// ProcessFluency 处理流畅度响应
func (p *ResponseProcessor) ProcessFluency(fluency *dto_evaluate.APIFluency, response *model.EvaluateResponse) {
	if fluency != nil {
		response.AIEvaluation.FluencyEvaluation.FluencyDescription = fluency.Comment
		response.AIEvaluation.FluencyEvaluation.FluencyScore = fluency.Score
	}
}

// ProcessOverall 处理总体评价响应
func (p *ResponseProcessor) ProcessOverall(overall *dto_evaluate.APIOverall, response *model.EvaluateResponse) {
	if overall != nil {
		response.AIEvaluation.OverallEvaluation.Description = overall.Comment
		response.AIEvaluation.OverallEvaluation.TopicRelevanceScore = overall.Score
	}
}

// ProcessExpression 处理表达评估响应
func (p *ResponseProcessor) ProcessExpression(expression *dto_evaluate.APIExpression, response *model.EvaluateResponse) {
	if expression != nil {
		response.AIEvaluation.ExpressionEvaluation.ExpressDescription = expression.Comment
		response.AIEvaluation.ExpressionEvaluation.ExpressionScore = expression.Score
	}
}

// ProcessSuggestion 处理建议响应
func (p *ResponseProcessor) ProcessSuggestion(suggestion *dto_evaluate.APISuggestion, response *model.EvaluateResponse) {
	if suggestion != nil {
		response.AIEvaluation.SuggestionEvaluation.SuggestionDescription = suggestion.Comment
	}
}

// ProcessParagraph 处理段落评估响应
func (p *ResponseProcessor) ProcessParagraph(paragraph *dto_evaluate.APIParagraph, response *model.EvaluateResponse) {
	if paragraph != nil {
		for i, comment := range paragraph.Comments {
			paragraphEval := model.ParagraphEvaluation{
				ParagraphIndex: i,
				Comment:        comment,
			}
			response.AIEvaluation.ParagraphEvaluations = append(response.AIEvaluation.ParagraphEvaluations, paragraphEval)
		}
	}
}

// ProcessScore 处理评分响应
func (p *ResponseProcessor) ProcessScore(score *model.APIScore, req *model.EvaluateRequest, response *model.EvaluateResponse) {
	if score == nil {
		return
	}

	logrus.Infof("response.EssayInfo.AllScore:%v, score.Result.Scores:%+v", response.EssayInfo.AllScore, score.Result.Scores)
	if response.EssayInfo.AllScore <= 0 {
		response.EssayInfo.AllScore = 100
	}

	// 设置评论
	response.AIEvaluation.ScoreEvaluation.Comment = score.Result.Comment
	response.AIEvaluation.ScoreEvaluation.Comments.Appearance = score.Result.Comments.Appearance
	response.AIEvaluation.ScoreEvaluation.Comments.Content = score.Result.Comments.Content
	response.AIEvaluation.ScoreEvaluation.Comments.Expression = score.Result.Comments.Expression
	response.AIEvaluation.ScoreEvaluation.Comments.Structure = score.Result.Comments.Structure
	response.AIEvaluation.ScoreEvaluation.Comments.Development = score.Result.Comments.Development

	// 设置原始分数
	response.AIEvaluation.ScoreEvaluation.Scores.All = score.Result.Scores.All
	response.AIEvaluation.ScoreEvaluation.Scores.Appearance = score.Result.Scores.Appearance
	response.AIEvaluation.ScoreEvaluation.Scores.Content = score.Result.Scores.Content
	response.AIEvaluation.ScoreEvaluation.Scores.Expression = score.Result.Scores.Expression
	response.AIEvaluation.ScoreEvaluation.Scores.Structure = score.Result.Scores.Structure
	response.AIEvaluation.ScoreEvaluation.Scores.Development = score.Result.Scores.Development

	// 计算总分比例（直接用上游分数和总分）
	response.AIEvaluation.ScoreEvaluation.Scores.AllWithTotal = fmt.Sprintf("%d/%d",
		score.Result.Scores.All, response.EssayInfo.AllScore)

	// 计算各项分数比例（使用 req 中的分项总分作为分母）
	if req.ContentScore != nil && *req.ContentScore > 0 {
		response.AIEvaluation.ScoreEvaluation.Scores.ContentWithTotal = fmt.Sprintf("%d/%d",
			score.Result.Scores.Content, *req.ContentScore)
	}

	if req.ExpressionScore != nil && *req.ExpressionScore > 0 {
		response.AIEvaluation.ScoreEvaluation.Scores.ExpressionWithTotal = fmt.Sprintf("%d/%d",
			score.Result.Scores.Expression, *req.ExpressionScore)
	}

	if req.StructureScore != nil && *req.StructureScore > 0 {
		response.AIEvaluation.ScoreEvaluation.Scores.StructureWithTotal = fmt.Sprintf("%d/%d",
			score.Result.Scores.Structure, *req.StructureScore)
	}

	if req.DevelopmentScore != nil && *req.DevelopmentScore > 0 {
		response.AIEvaluation.ScoreEvaluation.Scores.DevelopmentWithTotal = fmt.Sprintf("%d/%d",
			score.Result.Scores.Development, *req.DevelopmentScore)
	}
}

// ProcessPolishing 处理润色响应
func (p *ResponseProcessor) ProcessPolishing(polishing model.APIPolishingContent, response *model.EvaluateResponse) error {
	var paragraphEval model.PolishingEvaluation
	pIndex := polishing.ParagraphIdx
	paragraphEval.ParagraphIndex = pIndex

	for _, sentence := range polishing.Content {
		for _, edit := range sentence.Edits {
			pe := new(model.PolishingEdit)
			if err := copier.Copy(pe, edit); err != nil {
				logrus.Errorf("copy polishing edit failed, err:%v", err)
				continue
			}

			if pIndex >= len(response.Text) || pIndex < 0 {
				logrus.Errorf("出界啦!!!! pIndex:%v", pIndex)
				continue
			}

			_, pe.SentenceIndex, _ = lo.FindIndexOf(response.Text[pIndex], func(row string) bool {
				return strings.Contains(row, sentence.OriginalSentence)
			})
			if pe.SentenceIndex == -1 {
				logrus.Errorf("original未找到:%s, row:%+v, pIndex:%v", sentence.OriginalSentence, response.Text[pIndex], pIndex)
				continue
			}

			switch edit.Op {
			case "insert":
				pe.Original = edit.PositionAfter
				pe.Revised = edit.Text
			case "replace", "delete":
				pe.Original = edit.Original
				pe.Revised = edit.Replacement
			default:
				logrus.Errorf("未知操作类型, edit:%+v", edit)
				continue
			}

			if pe.SentenceIndex >= len(response.Text[pIndex]) || pe.SentenceIndex < 0 {
				logrus.Errorf("出界啦!!!! pIndex:%v, pe.SentenceIndex:%v", pIndex, pe.SentenceIndex)
				continue
			}

			originSentence := response.Text[pIndex][pe.SentenceIndex]

			index := strings.Index(originSentence, pe.Original)
			if index == -1 {
				logrus.Errorf("原句未找到:%s, row:%+v, originSentence:%v", pe.Original, response.Text[pIndex], originSentence)
				continue
			}
			beg := utf8.RuneCountInString(originSentence[:strings.Index(originSentence, pe.Original)])

			pe.Span = []int{beg + 1, beg + utf8.RuneCountInString(pe.Original)}

			paragraphEval.Edits = append(paragraphEval.Edits, *pe)
		}
	}

	response.AIEvaluation.PolishingEvaluation = append(response.AIEvaluation.PolishingEvaluation, paragraphEval)

	return nil
}

// ProcessEssayInfo 处理作文基本信息响应
func (p *ResponseProcessor) ProcessEssayInfo(essayInfo *dto_evaluate.APIEssayInfo, req *model.EvaluateRequest, response *model.EvaluateResponse) {
	if essayInfo == nil {
		return
	}

	if req.Grade != nil {
		essayInfo.Grade = *req.Grade
	}
	if req.EssayType != nil {
		essayInfo.EssayType = *req.EssayType
	}
	if req.TotalScore != nil {
		essayInfo.AllScore = *req.TotalScore
	}

	// 构建响应基础结构
	response.Title = req.Title
	response.Text = essayInfo.Sents
	response.EssayInfo = model.EssayInfo{
		EssayType: essayInfo.EssayType,
		Grade:     essayInfo.Grade,
		Counting: model.Counting{
			AdjAdvNum:         essayInfo.Counting.AdjAdvNum,
			CharNum:           essayInfo.Counting.CharNum,
			DieciNum:          essayInfo.Counting.DieciNum,
			Fluency:           essayInfo.Counting.Fluency,
			GrammarMistakeNum: essayInfo.Counting.GrammarMistakeNum,
			HighlightSentsNum: essayInfo.Counting.HighlightSentsNum,
			IdiomNum:          essayInfo.Counting.IdiomNum,
			NounTypeNum:       essayInfo.Counting.NounTypeNum,
			ParaNum:           essayInfo.Counting.ParaNum,
			SentNum:           essayInfo.Counting.SentNum,
			UniqueWordNum:     essayInfo.Counting.UniqueWordNum,
			VerbTypeNum:       essayInfo.Counting.VerbTypeNum,
			WordNum:           essayInfo.Counting.WordNum,
			WrittenMistakeNum: essayInfo.Counting.WrittenMistakeNum,
		},
		AllScore: essayInfo.AllScore,
	}
}

// InitializeResponse 初始化响应结构
func (p *ResponseProcessor) InitializeResponse(response *model.EvaluateResponse, modelVersion model.ModelVersion) {
	response.AIEvaluation = model.AIEvaluation{
		ModelVersion:           modelVersion,
		OverallEvaluation:      model.OverallEvaluation{},
		FluencyEvaluation:      model.FluencyEvaluation{},
		WordSentenceEvaluation: model.WordSentenceEvaluation{},
		ExpressionEvaluation:   model.ExpressionEvaluation{},
		SuggestionEvaluation:   model.SuggestionEvaluation{},
		ParagraphEvaluations:   []model.ParagraphEvaluation{},
		ScoreEvaluation:        model.ScoreEvaluation{},
		PolishingEvaluation:    []model.PolishingEvaluation{},
	}

	// 初始化好词好句评估结果
	sentencesEvaluations := make([][]model.SentenceEvaluation, len(response.Text))
	for i, paragraph := range response.Text {
		sentencesEvaluations[i] = make([]model.SentenceEvaluation, len(paragraph))
		for j := range paragraph {
			sentencesEvaluations[i][j] = model.SentenceEvaluation{
				IsGoodSentence:  false,
				Label:           "",
				Type:            make(map[string]string),
				WordEvaluations: []model.WordEvaluation{},
			}
		}
	}
	response.AIEvaluation.WordSentenceEvaluation.SentenceEvaluations = sentencesEvaluations
}

// ConvertToMap 将请求转换为map格式（用于API调用）
func (p *ResponseProcessor) ConvertToMap(title, content string, grade int, essayType string) map[string]any {
	return map[string]any{
		"title": title,
		"essay": content,
		"grade": grade,
		"type":  essayType,
	}
}

// SplitParagraphs 将文本分割为段落
func (p *ResponseProcessor) SplitParagraphs(content string) []string {
	content = p.cleaner.Clean(content)
	paragraphs := strings.Split(content, "\n")

	var result []string
	for _, para := range paragraphs {
		if trimmed := strings.TrimSpace(para); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// SplitSentences 将段落分割为句子
func (p *ResponseProcessor) SplitSentences(paragraph string) []string {
	// 简单的句子分割逻辑，可以根据需要改进
	paragraph = strings.TrimSpace(paragraph)
	if paragraph == "" {
		return []string{}
	}

	// 按句号、问号、感叹号分割
	var sentences []string
	var current strings.Builder

	for _, r := range paragraph {
		current.WriteRune(r)
		if r == '。' || r == '！' || r == '？' || r == '.' || r == '!' || r == '?' {
			if s := strings.TrimSpace(current.String()); s != "" {
				sentences = append(sentences, s)
			}
			current.Reset()
		}
	}

	// 添加剩余内容
	if s := strings.TrimSpace(current.String()); s != "" {
		sentences = append(sentences, s)
	}

	return sentences
}
