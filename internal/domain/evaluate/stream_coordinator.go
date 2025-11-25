package evaluate

import (
	"context"
	"encoding/json"
	dto_evaluate "essay-stateless/internal/dto/evaluate"
	"essay-stateless/internal/model"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// APIResult API调用结果
type APIResult struct {
	Step string // API步骤名
	Data any    // API返回数据
	Err  error  // 错误信息
}

// StreamCoordinator 流式处理协调器
type StreamCoordinator struct {
	retryExecutor     *RetryExecutor
	responseProcessor *ResponseProcessor
}

// NewStreamCoordinator 创建流式协调器
func NewStreamCoordinator() *StreamCoordinator {
	return &StreamCoordinator{
		retryExecutor:     NewRetryExecutor(DefaultRetryConfig()),
		responseProcessor: NewResponseProcessor(),
	}
}

// CoordinateEvaluation 协调评估流程
func (c *StreamCoordinator) CoordinateEvaluation(
	ctx context.Context,
	req *model.EvaluateRequest,
	resultChan chan<- *model.StreamEvaluateResponse,
	clients *APIClientsFactory,
	modelVersion model.ModelVersion,
) error {
	defer close(resultChan)

	// 发送初始化消息
	c.sendProgress(resultChan, "init", "开始作文批改", 0)

	essayInfoClient := clients.CreateEssayInfoClient()
	essayInfo, err := essayInfoClient.GetEssayInfo(ctx, req)
	if err != nil {
		resultChan <- &model.StreamEvaluateResponse{
			Type:      "error",
			Step:      "essay_info",
			Message:   "获取作文信息失败",
			Data:      &model.StreamErrorData{Error: err.Error(), Step: "essay_info"},
			Timestamp: time.Now().Unix(),
		}
		return err
	}

	// 构建响应结构
	response := &model.EvaluateResponse{}
	c.responseProcessor.ProcessEssayInfo(essayInfo, req, response)
	c.responseProcessor.InitializeResponse(response, modelVersion)

	// 发送作文信息完成消息
	c.sendProgress(resultChan, "essay_info", "作文信息分析完成", 15,
		&model.StreamInitData{Title: response.Title, Text: response.Text, EssayInfo: response.EssayInfo})

	apiResultChan := make(chan *APIResult, 9)
	var wg sync.WaitGroup
	wg.Add(9)

	essay := map[string]any{
		"title": req.Title,
		"essay": req.Content,
		"grade": req.Grade,
		"type":  req.EssayType,
	}

	go c.callAPIAsync(ctx, &wg, "word_sentence", func() (any, error) {
		return clients.CreateWordSentenceClient().Evaluate(ctx, essay)
	}, apiResultChan)

	go c.callAPIAsync(ctx, &wg, "grammar", func() (any, error) {
		return clients.CreateGrammarClient().Check(ctx, essay)
	}, apiResultChan)

	go c.callAPIAsync(ctx, &wg, "fluency", func() (any, error) {
		return clients.CreateFluencyClient().Evaluate(ctx, essay)
	}, apiResultChan)

	go c.callAPIAsync(ctx, &wg, "overall", func() (any, error) {
		return clients.CreateOverallClient().Evaluate(ctx, essay)
	}, apiResultChan)

	go c.callAPIAsync(ctx, &wg, "expression", func() (any, error) {
		return clients.CreateExpressionClient().Evaluate(ctx, essay)
	}, apiResultChan)

	go c.callAPIAsync(ctx, &wg, "suggestion", func() (any, error) {
		return clients.CreateSuggestionClient().Generate(ctx, essay)
	}, apiResultChan)

	go c.callAPIAsync(ctx, &wg, "paragraph", func() (any, error) {
		return clients.CreateParagraphClient().Evaluate(ctx, essay)
	}, apiResultChan)

	go c.callAPIAsync(ctx, &wg, "score", func() (any, error) {
		return clients.CreateScoreClient().Calculate(ctx, essay, req)
	}, apiResultChan)

	// 润色流式处理（特殊处理）
	go c.callPolishingStreamAsync(ctx, &wg, clients, essay, apiResultChan, response)

	go func() {
		wg.Wait()
		close(apiResultChan)
	}()

	c.aggregateResultsRealtime(response, resultChan, apiResultChan, req)

	// 发送完成消息
	c.sendComplete(resultChan, response)

	return nil
}

// callAPIAsync 异步调用API，完成后立即发送结果
func (c *StreamCoordinator) callAPIAsync(
	ctx context.Context,
	wg *sync.WaitGroup,
	stepName string,
	apiFunc func() (any, error),
	resultChan chan<- *APIResult,
) {
	defer wg.Done()

	startTime := time.Now()
	var result any
	var err error

	err = c.retryExecutor.Execute(ctx, func() error {
		result, err = apiFunc()
		return err
	}, stepName)

	elapsed := time.Since(startTime)

	if err != nil {
		logrus.Errorf("API调用失败 [%s] 耗时: %v, 错误: %v", stepName, elapsed, err)
	} else {
		logrus.Infof("API调用成功 [%s] 耗时: %v", stepName, elapsed)
	}

	resultChan <- &APIResult{
		Step: stepName,
		Data: result,
		Err:  err,
	}
}

// callPolishingStreamAsync 异步调用润色API（流式处理）
func (c *StreamCoordinator) callPolishingStreamAsync(
	ctx context.Context,
	wg *sync.WaitGroup,
	clients *APIClientsFactory,
	essay map[string]any,
	apiResultChan chan<- *APIResult,
	response *model.EvaluateResponse,
) {
	defer wg.Done()

	startTime := time.Now()
	streamChan := make(chan string, 10)
	errCh := make(chan error, 1)
	processedAny := false

	// 启动流式请求
	go func() {
		defer close(streamChan)
		polishingClient := clients.CreatePolishingClient()
		if err := polishingClient.PolishStream(ctx, essay, streamChan); err != nil {
			logrus.Errorf("流式润色API调用失败: %v", err)
			errCh <- err
		}
	}()

	// 处理流式数据
	for {
		select {
		case content, ok := <-streamChan:
			if !ok {
				// channel已关闭，发送完成信号
				goto polishingDone
			}

			// 解析润色内容
			var polishing model.APIPolishingContent
			if err := json.Unmarshal([]byte(content), &polishing); err != nil {
				logrus.Errorf("解析润色内容失败: %v, content: %s", err, content)
				continue
			}

			// 立即处理润色内容
			if err := c.responseProcessor.ProcessPolishing(polishing, response); err != nil {
				logrus.Errorf("处理润色内容失败: %v", err)
				continue
			}

			processedAny = true
			logrus.Infof("处理润色段落 %d 完成", polishing.ParagraphIdx)

		case streamErr := <-errCh:
			logrus.Errorf("润色流式处理错误: %v", streamErr)
			apiResultChan <- &APIResult{
				Step: "polishing",
				Data: nil,
				Err:  streamErr,
			}
			return

		case <-ctx.Done():
			logrus.Warn("润色处理被取消")
			apiResultChan <- &APIResult{
				Step: "polishing",
				Data: nil,
				Err:  ctx.Err(),
			}
			return
		}
	}

polishingDone:
	elapsed := time.Since(startTime)
	if processedAny {
		logrus.Infof("润色流式处理完成 耗时: %v", elapsed)
	} else {
		logrus.Warn("未处理任何润色内容")
	}

	// 发送完成信号
	apiResultChan <- &APIResult{
		Step: "polishing",
		Data: "done",
		Err:  nil,
	}
}

// sendProgress 发送进度消息
func (c *StreamCoordinator) sendProgress(ch chan<- *model.StreamEvaluateResponse, step, message string, progress int, data ...any) {
	var progressData any
	if len(data) > 0 {
		progressData = data[0]
	}

	select {
	case ch <- &model.StreamEvaluateResponse{
		Type:      "progress",
		Step:      step,
		Progress:  progress,
		Message:   message,
		Data:      progressData,
		Timestamp: time.Now().Unix(),
	}:
	default:
		logrus.Warn("进度channel已满，跳过消息")
	}
}

// sendComplete 发送完成消息
func (c *StreamCoordinator) sendComplete(ch chan<- *model.StreamEvaluateResponse, data *model.EvaluateResponse) {
	ch <- &model.StreamEvaluateResponse{
		Type:      "complete",
		Step:      "finish",
		Progress:  100,
		Message:   "作文批改完成",
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// aggregateResultsRealtime 实时聚合处理（谁先完成谁先处理，动态progress）
func (c *StreamCoordinator) aggregateResultsRealtime(
	response *model.EvaluateResponse,
	progressChan chan<- *model.StreamEvaluateResponse,
	apiResultChan <-chan *APIResult,
	req *model.EvaluateRequest,
) {
	const totalAPIs = 9
	const baseProgress = 15  // essay_info完成后的进度
	const progressRange = 75 // 从15到90的范围

	completedCount := 0
	var errors []error

	// 实时监听API完成结果
	for result := range apiResultChan {
		completedCount++

		// 动态计算progress：谁先完成谁的progress就小
		currentProgress := baseProgress + int(float64(completedCount)/float64(totalAPIs)*float64(progressRange))

		if result.Err != nil {
			logrus.Errorf("API [%s] 执行失败: %v", result.Step, result.Err)
			errors = append(errors, result.Err)
			continue
		}

		// 根据step类型处理数据并发送进度
		c.processAndSendProgress(result, response, progressChan, currentProgress, req)

		logrus.Infof("进度更新: [%s] %d%% (%d/%d 完成)", result.Step, currentProgress, completedCount, totalAPIs)
	}

	if len(errors) > 0 {
		logrus.Errorf("共有 %d 个API调用失败，错误: %v", len(errors), errors)
	}

	logrus.Info("所有API结果处理完成！")
}

// processAndSendProgress 处理单个API结果并发送进度消息
func (c *StreamCoordinator) processAndSendProgress(
	result *APIResult,
	response *model.EvaluateResponse,
	progressChan chan<- *model.StreamEvaluateResponse,
	progress int,
	req *model.EvaluateRequest,
) {
	if result.Data == nil {
		logrus.Warnf("API [%s] 返回数据为空", result.Step)
		return
	}

	var stepData any

	switch result.Step {
	case "word_sentence":
		if wordSentence, ok := result.Data.(*dto_evaluate.APIWordSentence); ok {
			c.responseProcessor.ProcessWordSentence(wordSentence, response)
			stepData = model.AIEvaluation{WordSentenceEvaluation: response.AIEvaluation.WordSentenceEvaluation}
		}

	case "grammar":
		if grammar, ok := result.Data.(*dto_evaluate.APIGrammarInfo); ok {
			c.responseProcessor.ProcessGrammar(grammar, response)
			stepData = model.AIEvaluation{WordSentenceEvaluation: response.AIEvaluation.WordSentenceEvaluation}
		}

	case "fluency":
		if fluency, ok := result.Data.(*dto_evaluate.APIFluency); ok {
			c.responseProcessor.ProcessFluency(fluency, response)
			stepData = model.AIEvaluation{FluencyEvaluation: response.AIEvaluation.FluencyEvaluation}
		}

	case "overall":
		if overall, ok := result.Data.(*dto_evaluate.APIOverall); ok {
			c.responseProcessor.ProcessOverall(overall, response)
			stepData = model.AIEvaluation{OverallEvaluation: response.AIEvaluation.OverallEvaluation}
		}

	case "expression":
		if expression, ok := result.Data.(*dto_evaluate.APIExpression); ok {
			c.responseProcessor.ProcessExpression(expression, response)
			stepData = model.AIEvaluation{ExpressionEvaluation: response.AIEvaluation.ExpressionEvaluation}
		}

	case "suggestion":
		if suggestion, ok := result.Data.(*dto_evaluate.APISuggestion); ok {
			c.responseProcessor.ProcessSuggestion(suggestion, response)
			stepData = model.AIEvaluation{SuggestionEvaluation: response.AIEvaluation.SuggestionEvaluation}
		}

	case "paragraph":
		if paragraph, ok := result.Data.(*dto_evaluate.APIParagraph); ok {
			c.responseProcessor.ProcessParagraph(paragraph, response)
			stepData = model.AIEvaluation{ParagraphEvaluations: response.AIEvaluation.ParagraphEvaluations}
		}

	case "score":
		if score, ok := result.Data.(*model.APIScore); ok {
			c.responseProcessor.ProcessScore(score, req, response)
			stepData = model.AIEvaluation{ScoreEvaluation: response.AIEvaluation.ScoreEvaluation}
		}

	case "polishing":
		// 润色已经在流式处理中实时更新了
		stepData = model.AIEvaluation{PolishingEvaluation: response.AIEvaluation.PolishingEvaluation}

	default:
		logrus.Warnf("未知的API步骤: %s", result.Step)
		return
	}

	// 发送进度消息
	c.sendProgress(progressChan, result.Step, getStepMessage(result.Step), progress, stepData)
}

// getStepMessage 获取步骤的提示消息
func getStepMessage(step string) string {
	messages := map[string]string{
		"word_sentence": "词句评估完成",
		"grammar":       "语法检查完成",
		"fluency":       "流畅度评估完成",
		"overall":       "总体评价完成",
		"expression":    "表达评估完成",
		"suggestion":    "建议生成完成",
		"paragraph":     "段落评估完成",
		"score":         "评分完成",
		"polishing":     "作文润色完成",
	}
	if msg, ok := messages[step]; ok {
		return msg
	}
	return step + "完成"
}
