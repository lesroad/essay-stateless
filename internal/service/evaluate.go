package service

import (
	"context"
	"essay-stateless/internal/config"
	"essay-stateless/internal/model"
	"essay-stateless/pkg/httpclient"
	"sync"
	"time"
	"unicode/utf8"
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
	Evaluate(ctx context.Context, req *model.EvaluateRequest) (*model.EvaluateResponse, error)
	OcrEvaluate(ctx context.Context, req *model.OcrEvaluateRequest) (*model.EvaluateResponse, error)
	EvaluateStream(ctx context.Context, req *model.EvaluateRequest, ch chan<- *model.StreamEvaluateResponse) error // 单向发送通道
	OcrEvaluateStream(ctx context.Context, req *model.OcrEvaluateRequest, ch chan<- *model.StreamEvaluateResponse) error
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

func (s *evaluateService) Evaluate(ctx context.Context, req *model.EvaluateRequest) (*model.EvaluateResponse, error) {
	essay := map[string]interface{}{
		"title": req.Title,
		"essay": req.Content,
	}

	var essayInfo APIEssayInfo
	if err := s.httpClient.Post(ctx, s.config.API.EssayInfo, essay, &essayInfo); err != nil {
		return nil, err
	}

	if req.Grade != nil {
		essayInfo.Grade = *req.Grade
	}
	if req.EssayType != nil {
		essayInfo.EssayType = *req.EssayType
	}

	essay["grade"] = essayInfo.Grade
	essay["type"] = essayInfo.EssayType

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
	}

	return s.processAPIResponses(ctx, essay, response)
}

func (s *evaluateService) OcrEvaluate(ctx context.Context, req *model.OcrEvaluateRequest) (*model.EvaluateResponse, error) {
	provider := "bee"
	if req.Provider != nil {
		provider = *req.Provider
	}

	imageType := "url"
	if req.ImageType != nil {
		imageType = *req.ImageType
	}

	ocrReq := &model.TitleOcrRequest{
		Images:   req.Images,
		LeftType: &req.LeftType,
	}

	ocrResp, err := s.ocrService.TitleOcr(ctx, provider, imageType, ocrReq)
	if err != nil {
		return nil, err
	}

	evaluateReq := &model.EvaluateRequest{
		Title:     ocrResp.Title,
		Content:   ocrResp.Content,
		Grade:     req.Grade,
		EssayType: req.EssayType,
	}

	return s.Evaluate(ctx, evaluateReq)
}

func (s *evaluateService) processAPIResponses(ctx context.Context, essay map[string]interface{}, response *model.EvaluateResponse) (*model.EvaluateResponse, error) {
	var wg sync.WaitGroup
	wg.Add(8)

	var overall *APIOverall           // 总评
	var fluency *APIFluency           // 流畅度评语
	var wordSentence *APIWordSentence // 好词好句
	var grammarInfo *APIGrammarInfo   // 语法错误识别
	var expression *APIExpression     // 逻辑表达评语
	var suggestion *APISuggestion     // 修改建议评语
	var paragraph *APIParagraph       // 段落点评
	var score *APIScore               // 分数

	// 异步调用所有API
	go func() {
		defer wg.Done()
		s.httpClient.Post(ctx, s.config.API.Overall, essay, &overall)
	}()

	go func() {
		defer wg.Done()
		s.httpClient.Post(ctx, s.config.API.Fluency, essay, &fluency)
	}()

	go func() {
		defer wg.Done()
		s.httpClient.Post(ctx, s.config.API.WordSentence, essay, &wordSentence)
	}()

	go func() {
		defer wg.Done()
		s.httpClient.Post(ctx, s.config.API.GrammarInfo, essay, &grammarInfo)
	}()

	go func() {
		defer wg.Done()
		s.httpClient.Post(ctx, s.config.API.Expression, essay, &expression)
	}()

	go func() {
		defer wg.Done()
		s.httpClient.Post(ctx, s.config.API.Suggestion, essay, &suggestion)
	}()

	go func() {
		defer wg.Done()
		s.httpClient.Post(ctx, s.config.API.Paragraph, essay, &paragraph)
	}()

	go func() {
		defer wg.Done()
		essay["type"] = "essay"
		s.httpClient.Post(ctx, s.config.API.Score, essay, &score)
	}()

	wg.Wait()

	s.processWordSentence(wordSentence, response)
	s.processGrammarInfo(grammarInfo, response)
	s.processFluency(fluency, response)
	s.processParagraph(paragraph, response)
	s.processExpression(expression, response)
	s.processOverall(overall, response)
	s.processSuggestion(suggestion, response)
	s.processScore(score, response)

	return response, nil
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

func (s *evaluateService) processScore(score *APIScore, response *model.EvaluateResponse) {
	if score != nil {
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
	}
}

type APIEssayInfo struct {
	Grade     int              `json:"grade_int"`
	EssayType string           `json:"essay_type"`
	Counting  APIEssayCounting `json:"counting"`
	Sents     [][]string       `json:"sents"`
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

type APIScore struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Result  struct {
		Comment  string `json:"comment"`
		Comments struct {
			Appearance  string `json:"appearance"`
			Content     string `json:"content"`
			Expression  string `json:"expression"`
			Structure   string `json:"structure"`
			Development string `json:"development"`
		} `json:"comments"`
		Scores struct {
			All         int `json:"all"`
			Appearance  int `json:"appearance"`
			Content     int `json:"content"`
			Expression  int `json:"expression"`
			Structure   int `json:"structure"`
			Development int `json:"development"`
		} `json:"scores"`
	} `json:"result"`
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

	// 并发调用各种评估API，但逐步返回结果
	return s.processAPIResponsesStream(ctx, essay, response, ch)
}

// OcrEvaluateStream OCR流式批改评估
func (s *evaluateService) OcrEvaluateStream(ctx context.Context, req *model.OcrEvaluateRequest, ch chan<- *model.StreamEvaluateResponse) error {
	defer close(ch)

	// 发送初始化消息
	ch <- &model.StreamEvaluateResponse{
		Type:      "init",
		Step:      "ocr",
		Progress:  0,
		Message:   "开始OCR识别",
		Timestamp: time.Now().Unix(),
	}

	provider := "bee"
	if req.Provider != nil {
		provider = *req.Provider
	}

	imageType := "url"
	if req.ImageType != nil {
		imageType = *req.ImageType
	}

	ocrReq := &model.TitleOcrRequest{
		Images:   req.Images,
		LeftType: &req.LeftType,
	}

	ocrResp, err := s.ocrService.TitleOcr(ctx, provider, imageType, ocrReq)
	if err != nil {
		ch <- &model.StreamEvaluateResponse{
			Type:      "error",
			Step:      "ocr",
			Message:   "OCR识别失败",
			Data:      &model.StreamErrorData{Error: err.Error(), Step: "ocr"},
			Timestamp: time.Now().Unix(),
		}
		return err
	}

	// 发送OCR完成消息
	ch <- &model.StreamEvaluateResponse{
		Type:      "progress",
		Step:      "ocr",
		Progress:  10,
		Message:   "OCR识别完成",
		Data:      &model.StreamStepData{Step: "ocr", Data: ocrResp},
		Timestamp: time.Now().Unix(),
	}

	evaluateReq := &model.EvaluateRequest{
		Title:     ocrResp.Title,
		Content:   ocrResp.Content,
		Grade:     req.Grade,
		EssayType: req.EssayType,
	}

	return s.EvaluateStream(ctx, evaluateReq, ch)
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
				var score *APIScore
				scoreEssay := essay
				scoreEssay["prompt"] = ""
				scoreEssay["image"] = ""
				scoreEssay["type"] = "essay"
				err = s.httpClient.Post(ctx, step.URL, scoreEssay, &score)
				if err == nil {
					s.processScore(score, response)
					result = model.AIEvaluation{ScoreEvaluation: response.AIEvaluation.ScoreEvaluation}
				}
			}

			resultCh <- EvaluateResult{Step: step.Name, Message: step.Message, Data: result, Err: err}
		}(step)
	}

	// 收集所有结果并发送到输出通道
	for range steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case result := <-resultCh:
			if result.Err != nil {
				// 发送错误消息
				ch <- &model.StreamEvaluateResponse{
					Type:      "error",
					Step:      result.Step,
					Message:   getErrorMessage(result.Step),
					Data:      &model.StreamErrorData{Error: result.Err.Error(), Step: result.Step},
					Timestamp: time.Now().Unix(),
				}
			} else {
				// 发送进度消息
				ch <- &model.StreamEvaluateResponse{
					Type:      "progress",
					Step:      result.Step,
					Message:   result.Message,
					Data:      result.Data,
					Timestamp: time.Now().Unix(),
				}
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

// 获取错误消息的辅助函数
func getErrorMessage(step string) string {
	messages := map[string]string{
		"word_sentence": "好词好句分析失败",
		"grammar":       "语法检查失败",
		"fluency":       "流畅度评估失败",
		"overall":       "总体评价失败",
		"expression":    "表达评价失败",
		"suggestion":    "修改建议失败",
		"paragraph":     "段落点评失败",
		"score":         "打分失败",
	}
	return messages[step]
}
