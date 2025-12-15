package dto_evaluate

// 外部API响应类型定义
// 这些类型表示从外部评估服务返回的原始数据结构

// APIEssayInfo 作文基本信息响应
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

// APIOverall 总体评价响应
type APIOverall struct {
	Comment string `json:"comment"`
	Score   int    `json:"score"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// APIWordSentence 词句评估响应
type APIWordSentence struct {
	Data    APIWordSentenceData `json:"data"`
	Score   int                 `json:"score"`
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

// APIGrammarInfo 语法检查响应
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


// APISuggestion 建议响应
type APISuggestion struct {
	Comment string `json:"comment"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// APIParagraph 段落评估响应
type APIParagraph struct {
	Comments []string `json:"comments"`
	Code     string   `json:"code"`
	Message  string   `json:"message"`
}
