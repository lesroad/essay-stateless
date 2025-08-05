package model

type APIPolishingContent struct {
	ParagraphIdx int `json:"para_idx"`
	Content      []struct {
		OriginalSentence string             `json:"original_sentence"`
		Edits            []APIPolishingEdit `json:"edits"`
	} `json:"content"`
}

type APIPolishingEdit struct {
	Op            string `json:"op"`
	PositionAfter string `json:"position_after"`
	Text          string `json:"text"`
	Reason        string `json:"reason"`
	Original      string `json:"original"`
	Replacement   string `json:"replacement"`
}
