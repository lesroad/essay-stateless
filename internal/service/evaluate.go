package service

import (
	"context"
	"encoding/json"
	"essay-stateless/internal/config"
	"essay-stateless/internal/model"
	"essay-stateless/pkg/httpclient"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jinzhu/copier"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// EvaluateStep 评估步骤定义
type EvaluateStep struct {
	Name    string
	URL     string
	Message string
}

// EvaluateResult 评估结果
type EvaluateResult struct {
	Step    string
	Message string
	Data    interface{}
	Err     error
}

type EvaluateService interface {
	EvaluateStream(ctx context.Context, req *model.EvaluateRequest, ch chan<- *model.StreamEvaluateResponse) error // 单向发送通道
}

type evaluateService struct {
	config     *config.EvaluateConfig
	httpClient *httpclient.Client
	ocrService OcrService
}

func NewEvaluateService(config *config.EvaluateConfig, ocrService OcrService) EvaluateService {
	return &evaluateService{
		config:     config,
		httpClient: httpclient.New(),
		ocrService: ocrService,
	}
}

func (s *evaluateService) processWordSentence(wordSentence *APIWordSentence, response *model.EvaluateResponse) {
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

	for _, sent := range wordSentence.Data.Results.GoodSents {
		if sent.ParagraphID < len(sentencesEvaluations) && sent.SentID < len(sentencesEvaluations[sent.ParagraphID]) {
			sentenceEval := &sentencesEvaluations[sent.ParagraphID][sent.SentID]
			sentenceEval.IsGoodSentence = true
			sentenceEval.Label = sent.Label
			sentenceEval.Type["level1"] = "作文亮点"
			sentenceEval.Type["level2"] = "好句"
		}
	}

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

func (s *evaluateService) processGrammarInfo(grammarInfo *APIGrammarInfo, response *model.EvaluateResponse) {
	if grammarInfo == nil {
		return
	}
	time.Sleep(2 * time.Second)

	for _, typo := range grammarInfo.Grammar.Typo {
		gp := s.getSentenceRelativeIndex(response.Text, typo.StartPos)
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

type GrammarPosition struct {
	ParagraphIndex int
	SentenceIndex  int
	RelativeIndex  int
}

/*
// 假设文本结构如下：
text = [

	["第一段第一句", "第一段第二句"],
	["第二段第一句", "第二段第二句"]

]

// 如果 startPos = 8
// 函数会计算：
// - 第一段第一句：位置 0-4 (长度5)
// - 第一段第二句：位置 5-9 (长度5)
// startPos=8 落在第一段第二句中，相对位置是 8-5=3

// 返回结果：

	{
	    ParagraphIndex: 0,  // 第一段
	    SentenceIndex: 1,   // 第二句
	    RelativeIndex: 3    // 句子中第3个字符位置
	}
*/

// 字符偏移量
func (s *evaluateService) getSentenceRelativeIndex(text [][]string, startPos int) *GrammarPosition {
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

func (s *evaluateService) processFluency(fluency *APIFluency, response *model.EvaluateResponse) {
	if fluency != nil {
		response.AIEvaluation.FluencyEvaluation.FluencyDescription = fluency.Comment
		response.AIEvaluation.FluencyEvaluation.FluencyScore = fluency.Score
	}
}

func (s *evaluateService) processParagraph(paragraph *APIParagraph, response *model.EvaluateResponse) {
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

func (s *evaluateService) processExpression(expression *APIExpression, response *model.EvaluateResponse) {
	if expression != nil {
		response.AIEvaluation.ExpressionEvaluation.ExpressDescription = expression.Comment
		response.AIEvaluation.ExpressionEvaluation.ExpressionScore = expression.Score
	}
}

func (s *evaluateService) processSuggestion(suggestion *APISuggestion, response *model.EvaluateResponse) {
	if suggestion != nil {
		response.AIEvaluation.SuggestionEvaluation.SuggestionDescription = suggestion.Comment
	}
}

func (s *evaluateService) processOverall(overall *APIOverall, response *model.EvaluateResponse) {
	if overall != nil {
		response.AIEvaluation.OverallEvaluation.Description = overall.Comment
		response.AIEvaluation.OverallEvaluation.TopicRelevanceScore = overall.Score
	}
}

func (s *evaluateService) processScore(score *model.APIScore, response *model.EvaluateResponse) {
	if score != nil {
		// 分值转换
		logrus.Infof("response.EssayInfo.AllScore:%v, score.Result.Scores:%+v", response.EssayInfo.AllScore, score.Result.Scores)
		if response.EssayInfo.AllScore <= 0 {
			response.EssayInfo.AllScore = 100
		}

		response.AIEvaluation.ScoreEvaluation.Comment = score.Result.Comment
		response.AIEvaluation.ScoreEvaluation.Comments.Appearance = score.Result.Comments.Appearance
		response.AIEvaluation.ScoreEvaluation.Comments.Content = score.Result.Comments.Content
		response.AIEvaluation.ScoreEvaluation.Comments.Expression = score.Result.Comments.Expression
		response.AIEvaluation.ScoreEvaluation.Comments.Structure = score.Result.Comments.Structure
		response.AIEvaluation.ScoreEvaluation.Comments.Development = score.Result.Comments.Development
		response.AIEvaluation.ScoreEvaluation.Scores.All = score.Result.Scores.All
		response.AIEvaluation.ScoreEvaluation.Scores.Appearance = score.Result.Scores.Appearance
		response.AIEvaluation.ScoreEvaluation.Scores.Content = score.Result.Scores.Content
		response.AIEvaluation.ScoreEvaluation.Scores.Expression = score.Result.Scores.Expression
		response.AIEvaluation.ScoreEvaluation.Scores.Structure = score.Result.Scores.Structure
		response.AIEvaluation.ScoreEvaluation.Scores.Development = score.Result.Scores.Development

		// 最终分数/总分
		allScore := decimal.NewFromInt(score.Result.Scores.All).Div(decimal.NewFromInt(90)).Mul(decimal.NewFromInt(response.EssayInfo.AllScore)).Round(0).IntPart()
		response.AIEvaluation.ScoreEvaluation.Scores.AllWithTotal = fmt.Sprintf("%d/%d", allScore, response.EssayInfo.AllScore)

		// 内容分值/总分
		contentAllScore := DivideAndRoundUp(response.EssayInfo.AllScore, 3)
		contentScore := decimal.NewFromInt(score.Result.Scores.Content).Div(decimal.NewFromInt(30)).Mul(decimal.NewFromInt(contentAllScore)).Round(0).IntPart()
		response.AIEvaluation.ScoreEvaluation.Scores.ContentWithTotal = fmt.Sprintf("%d/%d", contentScore, contentAllScore)

		// 表达分值/总分
		expressionAllScore := DivideAndRoundDown(response.EssayInfo.AllScore, 3)
		expressionScore := decimal.NewFromInt(score.Result.Scores.Expression).Div(decimal.NewFromInt(30)).Mul(decimal.NewFromInt(expressionAllScore)).Round(0).IntPart()
		response.AIEvaluation.ScoreEvaluation.Scores.ExpressionWithTotal = fmt.Sprintf("%d/%d", expressionScore, expressionAllScore)

		if score.Result.Scores.Structure > 0 {
			// 结构分值/总分
			structureAllScore := DivideAndRoundDown(response.EssayInfo.AllScore, 3)
			structureScore := allScore - contentScore - expressionScore
			response.AIEvaluation.ScoreEvaluation.Scores.StructureWithTotal = fmt.Sprintf("%d/%d", structureScore, structureAllScore)
		} else {
			// 发展分值/总分
			developmentAllScore := DivideAndRoundDown(response.EssayInfo.AllScore, 3)
			developmentScore := allScore - contentScore - expressionScore
			response.AIEvaluation.ScoreEvaluation.Scores.DevelopmentWithTotal = fmt.Sprintf("%d/%d", developmentScore, developmentAllScore)
		}
	}
}

type APIEssayInfo struct {
	Grade     int              `json:"grade_int"`
	EssayType string           `json:"essay_type"`
	Counting  APIEssayCounting `json:"counting"`
	Sents     [][]string       `json:"sents"`
	AllScore  int64            `json:"score_int"`
	Code      string           `json:"code"`
	Message   string           `json:"message"`
}

type APIEssayCounting struct {
	AdjAdvNum         int `json:"adj_adv_num"`
	CharNum           int `json:"char_num"`
	DieciNum          int `json:"dieci_num"`
	Fluency           int `json:"fluency"`
	GrammarMistakeNum int `json:"grammar_mistake_num"`
	HighlightSentsNum int `json:"highlight_sents_num"`
	IdiomNum          int `json:"idiom_num"`
	NounTypeNum       int `json:"noun_type_num"`
	ParaNum           int `json:"para_num"`
	SentNum           int `json:"sent_num"`
	UniqueWordNum     int `json:"unique_word_num"`
	VerbTypeNum       int `json:"verb_type_num"`
	WordNum           int `json:"word_num"`
	WrittenMistakeNum int `json:"written_mistake_num"`
}

type APIOverall struct {
	Comment string `json:"comment"`
	Score   int    `json:"score"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIFluency struct {
	Comment string `json:"comment"`
	Score   int    `json:"score"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

/*
	{
	  "code": 200,
	  "data": {
	    "grade": 2,
	    "results": {
	      "good_sents": [ // 好句分析
	        {
	          "label": "排比", // 修辞手法
	          "paragraph_id": 0, // 第0段
	          "sent_id": 2  // 第2句
	        },
	        {
	          "label": "比拟",
	          "paragraph_id": 0,
	          "sent_id": 3
	        },
	        {
	          "label": "比拟",
	          "paragraph_id": 0,
	          "sent_id": 5
	        }
	      ],
	      "good_words": [ // 好词分析
	        {
	          "end": 11, // 结束位置：第11个字符，即：五颜六色
	          "paragraph_id": 0, // 第0段
	          "sent_id": 2, // 第2句
	          "start": 7  // 开始位置：第7个字符
	        }
	      ],
	      "type": "good_expression"
	    },
	    "sents": [
	      [ // 第0段（只有一个段落）
	        "春天来了，广场上十分热闹，孩子们都在放风筝。",  // 第0句
	        "\\n广场上，大家都在放风筝。",
	        "风筝各种各样，五颜六色，有小鸟的、蝴蝶的，金鱼的，十分美丽，广场上有两个孩子，一个在放蝴蝶风筝，一个在放小鸟风筝，还有一个在把着小鸟风筝的脚，把着小鸟风筝杆的人好似在说："你先把着，一会儿风来了你再放手！""",
	        "把着小鸟风筝脚的说："好！""",
	        "还有一对夫妻，一个孩子，小孩正在放着三角形的风筝，那对夫妻脸上挂着幸福的笑容，心里好像在想着什么？",
	        "还有一个老鹰，已经看不到是谁在放了，那条飞龙风筝也看不见了。",
	        "广场上的人脸上都挂着开心的笑容。" // 第6句
	      ]
	    ],
	    "type": "good_expression"
	  },
	  "message": "response success"
	}
*/
type APIWordSentence struct {
	Data    APIWordSentenceData `json:"data"`
	Score   int                 `json:"score"` // 貌似没值？
	Code    int                 `json:"code"`
	Message string              `json:"message"`
}

type APIWordSentenceData struct {
	Results APIWordSentenceResults `json:"results"`
}

type APIWordSentenceResults struct {
	GoodSents []APIGoodSent `json:"good_sents"`
	GoodWords []APIGoodWord `json:"good_words"`
}

type APIGoodSent struct {
	ParagraphID int    `json:"paragraph_id"`
	SentID      int    `json:"sent_id"`
	Label       string `json:"label"`
}

type APIGoodWord struct {
	ParagraphID int `json:"paragraph_id"`
	SentID      int `json:"sent_id"`
	Start       int `json:"start"`
	End         int `json:"end"`
}

/*
	{
	    "code": "200",
	    "grammar": {
	        "typo": [
	            {
	                "end_pos": 57,
	                "extra": "春天来了，广场上十分热闹，孩子们都在放风筝。\\n广场上，大家都在放风筝。风筝各种各样，五颜六色，有小鸟的、蝴蝶的，金鱼的，十分美丽，广场上有两个孩子，一个在放蝴蝶风筝，一个在放小鸟风筝，还有一个在把着小鸟风筝的脚，把着小鸟风筝杆的人好似在说："你先把着，一会儿风来了你再放手！""把着小鸟风筝脚的说：""好！""还有一对夫妻，一个孩子，小孩正在放着三角形的风筝，那对夫妻脸上挂着幸福的笑容，心里好像在想着什么？还有一个老鹰，已经看不到是谁在放了，那条飞龙风筝也看不见了。广场上的人脸上都挂着开心的笑容。",
	                "ori": "，",
	                "revised": "、",
	                "start_pos": 56,
	                "type": "标点问题"
	            },
	            {
	                "end_pos": 66,
	                "extra": "春天来了，广场上十分热闹，孩子们都在放风筝。\\n广场上，大家都在放风筝。风筝各种各样，五颜六色，有小鸟的、蝴蝶的，金鱼的，十分美丽，广场上有两个孩子，一个在放蝴蝶风筝，一个在放小鸟风筝，还有一个在把着小鸟风筝的脚，把着小鸟风筝杆的人好似在说："你先把着，一会儿风来了你再放手！""把着小鸟风筝脚的说：""好！""还有一对夫妻，一个孩子，小孩正在放着三角形的风筝，那对夫妻脸上挂着幸福的笑容，心里好像在想着什么？还有一个老鹰，已经看不到是谁在放了，那条飞龙风筝也看不见了。广场上的人脸上都挂着开心的笑容。",
	                "ori": "，",
	                "revised": "。",
	                "start_pos": 65,
	                "type": "标点问题"
	            }
	        ]
	    },
	    "message": "response success"
	}
*/
type APIGrammarInfo struct {
	Grammar APIGrammar `json:"grammar"`
	Code    string     `json:"code"`
	Message string     `json:"message"`
}

type APIGrammar struct {
	Typo []APITypo `json:"typo"`
}

type APITypo struct {
	StartPos int    `json:"start_pos"`
	EndPos   int    `json:"end_pos"`
	Type     string `json:"type"`
	Ori      string `json:"ori"`
	Revised  string `json:"revised"`
}

type APIExpression struct {
	Comment string `json:"comment"`
	Score   int    `json:"score"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APISuggestion struct {
	Comment string `json:"comment"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type APIParagraph struct {
	Comments []string `json:"comments"`
	Code     string   `json:"code"`
	Message  string   `json:"message"`
}

/*
	{
	    "type": "content",
	    "content": [
	        {
	            "original_sentence": "我的心爱之物是一个小娃娃，那是我爸爸出差的时候给我带的礼物。",
	            "edits": [
	                {
	                    "op": "replace",
	                    "original": "一个小娃娃",
	                    "replacement": "一尊绒布小娃娃",
	                    "reason": "增加材质描写“绒布”，使小娃娃的形象更具体可感"
	                },
	                {
	                    "op": "insert",
	                    "position_after": "出差的时候",
	                    "text": "特意",
	                    "reason": "“特意”一词体现了爸爸对“我”的用心，暗含情感温度，让礼物更显珍贵"
	                }
	            ]
	        },
	        {
	            "original_sentence": "她已经陪伴我5年半了。",
	            "edits": [
	                {
	                    "op": "insert",
	                    "position_after": "陪伴我",
	                    "text": "整整",
	                    "reason": "“整整”强调了时间的长久，突出小娃娃陪伴“我”的时间跨度，情感更浓厚"
	                }
	            ]
	        },
	        {
	            "original_sentence": "那她为什么是我的心爱之物呢？",
	            "edits": [
	                {
	                    "op": "replace",
	                    "original": "那她为什么是我的心爱之物呢？",
	                    "replacement": "你猜，她为什么能成为我的心爱之物呢？",
	                    "reason": "运用设问和第二人称“你”，拉近与读者的距离，引发阅读兴趣，使开头更活泼自然"
	                }
	            ]
	        }
	    ]
	}
*/

// 润色点评
func (s *evaluateService) processPolishing(polishing model.APIPolishingContent, response *model.EvaluateResponse) error {
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

// EvaluateStream 流式批改评估
func (s *evaluateService) EvaluateStream(ctx context.Context, req *model.EvaluateRequest, ch chan<- *model.StreamEvaluateResponse) error {
	defer close(ch)
	// 发送初始化消息
	ch <- &model.StreamEvaluateResponse{
		Type:      "init",
		Step:      "start",
		Progress:  0,
		Message:   "开始作文批改",
		Timestamp: time.Now().Unix(),
	}

	// 0. 调整作文格式，去除多余\n和非正常作文标点的特殊符号
	req.Content = cleanContent(req.Content)
	logrus.Infof("调整后作文：%s", req.Content)

	// 1. 获取作文基本信息
	essay := map[string]interface{}{
		"title": req.Title,
		"essay": req.Content,
	}

	var essayInfo APIEssayInfo
	if err := s.httpClient.Post(ctx, s.config.API.EssayInfo, essay, &essayInfo); err != nil {
		ch <- &model.StreamEvaluateResponse{
			Type:      "error",
			Step:      "essay_info",
			Message:   "获取作文信息失败",
			Data:      &model.StreamErrorData{Error: err.Error(), Step: "essay_info"},
			Timestamp: time.Now().Unix(),
		}
		return err
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

	essay["grade"] = essayInfo.Grade
	essay["type"] = essayInfo.EssayType

	// 构建响应基础结构
	response := &model.EvaluateResponse{}
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

	response.AIEvaluation = model.AIEvaluation{
		ModelVersion: model.ModelVersion{
			Name:    s.config.ModelVersion.Name,
			Version: s.config.ModelVersion.Version,
		},
		OverallEvaluation:      model.OverallEvaluation{},
		FluencyEvaluation:      model.FluencyEvaluation{},
		WordSentenceEvaluation: model.WordSentenceEvaluation{},
		ExpressionEvaluation:   model.ExpressionEvaluation{},
		SuggestionEvaluation:   model.SuggestionEvaluation{},
		ParagraphEvaluations:   []model.ParagraphEvaluation{},
		ScoreEvaluation:        model.ScoreEvaluation{},
		PolishingEvaluation:    []model.PolishingEvaluation{},
	}

	// 发送初始化完成消息
	ch <- &model.StreamEvaluateResponse{
		Type:      "progress",
		Step:      "essay_info",
		Progress:  15,
		Message:   "作文信息分析完成",
		Data:      &model.StreamInitData{Title: response.Title, Text: response.Text, EssayInfo: response.EssayInfo},
		Timestamp: time.Now().Unix(),
	}

	logrus.Infof("作文详情：%+v", response.Text)

	// 并发调用各种评估API，但逐步返回结果
	return s.processAPIResponsesStream(ctx, essay, response, ch)
}

// 流式处理API响应
func (s *evaluateService) processAPIResponsesStream(ctx context.Context, essay map[string]interface{}, response *model.EvaluateResponse, ch chan<- *model.StreamEvaluateResponse) error {
	// 定义评估步骤
	steps := []EvaluateStep{
		{"word_sentence", s.config.API.WordSentence, "好词好句分析完成"},
		{"grammar", s.config.API.GrammarInfo, "语法检查完成"},
		{"fluency", s.config.API.Fluency, "流畅度评估完成"},
		{"overall", s.config.API.Overall, "总体评价完成"},
		{"expression", s.config.API.Expression, "表达评价完成"},
		{"suggestion", s.config.API.Suggestion, "修改建议完成"},
		{"paragraph", s.config.API.Paragraph, "段落点评完成"},
		{"score", s.config.API.Score, "打分完成"},
		{"polishing", s.config.API.Polishing, "作文润色完成"},
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

	// 创建结果通道和错误通道
	resultCh := make(chan EvaluateResult, len(steps))

	// 并发执行所有评估步骤
	for _, step := range steps {
		go func(step EvaluateStep) {
			select {
			case <-ctx.Done():
				resultCh <- EvaluateResult{Step: step.Name, Err: ctx.Err()}
				return
			default:
			}

			var result interface{}
			var err error

			// 调用API
			switch step.Name {
			case "word_sentence":
				var wordSentence *APIWordSentence
				err = s.httpClient.Post(ctx, step.URL, essay, &wordSentence)
				if err == nil {
					s.processWordSentence(wordSentence, response)
					result = model.AIEvaluation{WordSentenceEvaluation: response.AIEvaluation.WordSentenceEvaluation}
				}

			case "grammar":
				var grammarInfo *APIGrammarInfo
				err = s.httpClient.Post(ctx, step.URL, essay, &grammarInfo)
				if err == nil {
					s.processGrammarInfo(grammarInfo, response)
					result = model.AIEvaluation{WordSentenceEvaluation: response.AIEvaluation.WordSentenceEvaluation}
				}

			case "fluency":
				var fluency *APIFluency
				err = s.httpClient.Post(ctx, step.URL, essay, &fluency)
				if err == nil {
					s.processFluency(fluency, response)
					result = model.AIEvaluation{FluencyEvaluation: response.AIEvaluation.FluencyEvaluation}
				}

			case "overall":
				var overall *APIOverall
				err = s.httpClient.Post(ctx, step.URL, essay, &overall)
				if err == nil {
					s.processOverall(overall, response)
					result = model.AIEvaluation{OverallEvaluation: response.AIEvaluation.OverallEvaluation}
				}

			case "expression":
				var expression *APIExpression
				err = s.httpClient.Post(ctx, step.URL, essay, &expression)
				if err == nil {
					s.processExpression(expression, response)
					result = model.AIEvaluation{ExpressionEvaluation: response.AIEvaluation.ExpressionEvaluation}
				}

			case "suggestion":
				var suggestion *APISuggestion
				err = s.httpClient.Post(ctx, step.URL, essay, &suggestion)
				if err == nil {
					s.processSuggestion(suggestion, response)
					result = model.AIEvaluation{SuggestionEvaluation: response.AIEvaluation.SuggestionEvaluation}
				}

			case "paragraph":
				var paragraph *APIParagraph
				err = s.httpClient.Post(ctx, step.URL, essay, &paragraph)
				if err == nil {
					s.processParagraph(paragraph, response)
					result = model.AIEvaluation{ParagraphEvaluations: response.AIEvaluation.ParagraphEvaluations}
				}

			case "score":
				var score *model.APIScore
				scoreEssay := essay
				scoreEssay["prompt"] = ""
				scoreEssay["image"] = ""
				scoreEssay["type"] = "essay"
				err = s.httpClient.Post(ctx, step.URL, scoreEssay, &score)
				if err == nil {
					s.processScore(score, response)
					result = model.AIEvaluation{ScoreEvaluation: response.AIEvaluation.ScoreEvaluation}
				}

			case "polishing":
				var subResultCh = make(chan string)
				var streamErr error
				var processedAny bool

				// 使用一个错误通道来传播 goroutine 中的错误
				errCh := make(chan error, 1)

				go func() {
					defer close(subResultCh)
					defer close(errCh)
					if streamErr := s.httpClient.PostWithStream(ctx, step.URL, nil, essay, subResultCh); streamErr != nil {
						logrus.WithError(streamErr).Error("润色API调用失败")
						errCh <- streamErr
					}
				}()

				// 使用 select 来同时监听数据和错误
				for {
					select {
					case content, ok := <-subResultCh:
						if !ok {
							// channel 已关闭，检查是否有错误
							select {
							case streamErr = <-errCh:
								// 有错误
							default:
								// 没有错误，正常结束
							}
							goto polishingDone
						}

						polishing := model.APIPolishingContent{}
						if parseErr := json.Unmarshal([]byte(content), &polishing); parseErr != nil {
							logrus.WithError(parseErr).WithField("content", content).Error("解析润色内容失败")
							continue
						}

						if processErr := s.processPolishing(polishing, response); processErr != nil {
							logrus.WithError(processErr).Error("处理润色内容失败")
							continue
						}

						processedAny = true

					case streamErr = <-errCh:
						// 收到错误，退出循环
						goto polishingDone

					case <-ctx.Done():
						// 上下文取消
						streamErr = ctx.Err()
						goto polishingDone
					}
				}

			polishingDone:
				// 如果没有处理任何内容且没有明确的错误，这也是一个问题
				if !processedAny && streamErr == nil {
					streamErr = fmt.Errorf("润色API没有返回任何内容")
				}

				result = model.AIEvaluation{PolishingEvaluation: response.AIEvaluation.PolishingEvaluation}
				err = streamErr
			}

			resultCh <- EvaluateResult{Step: step.Name, Message: step.Message, Data: result, Err: err}
		}(step)
	}

	for range steps {
		result := <-resultCh
		if result.Err != nil {
			logrus.Errorf("评估步骤 %s 失败: %v", result.Step, result.Err)
		} else {
			ch <- &model.StreamEvaluateResponse{
				Type:      "progress",
				Step:      result.Step,
				Message:   result.Message,
				Data:      result.Data,
				Timestamp: time.Now().Unix(),
			}
		}
	}

	// 发送完成消息
	ch <- &model.StreamEvaluateResponse{
		Type:      "complete",
		Step:      "finish",
		Progress:  100,
		Message:   "作文批改完成",
		Data:      response,
		Timestamp: time.Now().Unix(),
	}

	return nil
}

// cleanContent 清理作文内容中的多余换行符和特殊符号
func cleanContent(content string) string {
	if content == "" {
		return content
	}

	// 1. 使用正则表达式将连续的多个\n替换为单个\n
	re := regexp.MustCompile(`\n+`)
	content = re.ReplaceAllString(content, "\n")

	// 2. 去除开头和结尾的换行符
	content = strings.Trim(content, "\n")

	// 3. 清理非正常作文标点的特殊符号
	// 保留正常的中文标点符号：。，！？；：""''（）【】《》、
	// 保留正常的英文标点符号：.,!?;:"'()-
	// 保留数字、中英文字母、空格、换行符

	// 定义需要保留的字符模式
	// \p{Han}: 中文字符
	// a-zA-Z0-9: 英文字母和数字
	// \s: 空白字符（包括空格、换行等）
	// 中文标点：。，！？；：""''（）【】《》、
	// 英文标点：.,!?;:"'()-
	validChars := regexp.MustCompile(`[^\p{Han}a-zA-Z0-9\s。，！？；：""''（）【】《》、.,!?;:"'()\-]`)
	content = validChars.ReplaceAllString(content, "")

	// 4. 清理多余的空格（连续多个空格替换为单个空格）
	spaceRe := regexp.MustCompile(`[ \t]+`)
	content = spaceRe.ReplaceAllString(content, " ")

	// 5. 去除行首行尾的空格
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	content = strings.Join(lines, "\n")

	// 6. 再次去除开头和结尾的换行符（防止清理后产生的多余换行）
	content = strings.Trim(content, "\n")

	return content
}
