package sales

import "fmt"

// Language represents the supported languages
type Language string

const (
	LanguageEnglish Language = "en"
	LanguageChinese Language = "zh"
)

// SalesScenario represents a sales training scenario
type SalesScenario struct {
	Product      string   `json:"product"`
	CustomerType string   `json:"customerType"`
	Goal         string   `json:"goal"`
	Language     Language `json:"language"`
}

// DialogueLine represents a single line in the dialogue
type DialogueLine struct {
	Speaker    string `json:"speaker"`    // "Salesperson" or "Customer"
	Content    string `json:"content"`    // Dialogue content
	Psychology string `json:"psychology"` // Psychology/subtext
	Highlight  string `json:"highlight"`  // highlight type: "skill", "concern", "turning", ""
}

// AnalysisPoint represents a point in the analysis
type AnalysisPoint struct {
	Title       string `json:"title"`       // Title
	DialogueRef string `json:"dialogueRef"` // Related dialogue segment
	Explanation string `json:"explanation"` // Explanation
}

// Score represents the sales skill scores
type Score struct {
	Opening           int `json:"opening"`           // Opening & ice-breaking
	NeedsDiscovery    int `json:"needsDiscovery"`    // Needs discovery
	ProductIntro      int `json:"productIntro"`      // Product introduction
	ObjectionHandling int `json:"objectionHandling"` // Objection handling
	Closing           int `json:"closing"`           // Closing
	Overall           int `json:"overall"`           // Overall performance
}

// SalesResult represents the complete sales training result
type SalesResult struct {
	Scenario     SalesScenario   `json:"scenario"`
	Dialogue     []DialogueLine  `json:"dialogue"`
	Success      bool            `json:"success"`      // Whether closed
	Strengths    []AnalysisPoint `json:"strengths"`    // What went well
	Improvements []AnalysisPoint `json:"improvements"` // Areas for improvement
	Scores       Score           `json:"scores"`       // Scores
	KeyLearning  string          `json:"keyLearning"`  // Key takeaways
}

// DefaultScenario returns the default gym membership scenario
func DefaultScenario() SalesScenario {
	return SalesScenario{
		Product:      "Gym Annual Membership",
		CustomerType: "Hesitant",
		Goal:         "Close the annual membership sale, emphasize continuous improvement",
		Language:     LanguageEnglish,
	}
}

// GenerateSystemPrompt returns the system-level prompt (role definition + format instructions)
func GenerateSystemPrompt(lang Language) string {
	if lang == LanguageChinese {
		return `你是一位專業的銷售培訓師和對話分析專家。

你的任務是生成銷售電話對話並進行詳細分析。

## 對話要求
1. 對話需展示完整的銷售流程：開場、需求探索、產品介紹、處理異議、促成交易、結束
2. 包含真實的客戶疑慮（沒時間、沒動力、怕浪費錢、怕堅持不了）
3. 銷售員需運用適當的銷售技巧（SPIN銷售法、FAB法則、同理心等）
4. 重點展示「每次進步一點點」的概念，讓客戶感受到成長的可能性
5. 結局隨機選擇成功或失敗

## 輸出格式（嚴格按照此格式）

### 對話內容
[銷售員]：{對話內容}
[心理]：{銷售員的策略意圖}
[客人]：{對話內容}
[心理]：{客人的真實想法}
...（繼續對話）

### 結果
{成功成交/交易失敗}

### 分析

#### 做得好的地方
1. **{技巧名稱}**
   - 對話：「{相關對話段落}」
   - 分析：{為什麼這樣做有效}

#### 可改進的地方
1. **{問題描述}**
   - 對話：「{相關對話段落}」
   - 建議：{具體改進建議}

#### 評分（1-10分）
- 開場與破冰：X分
- 需求探索：X分
- 產品介紹：X分
- 異議處理：X分
- 促成交易：X分
- 整體表現：X分

#### 關鍵學習點
{最重要的收穫和建議}

請確保對話自然流暢，分析具體且有建設性。`
	}
	return `You are a professional sales trainer and dialogue analyst.

Your task is to generate sales phone dialogues and provide detailed analysis.

## Dialogue Requirements
1. Show complete sales process: opening, needs discovery, product introduction, objection handling, closing, follow-up
2. Include realistic customer concerns (no time, no motivation, cost concerns, fear of quitting)
3. Use appropriate sales techniques (SPIN selling, FAB model, empathy, etc.)
4. Emphasize the concept of "small progress every day"
5. Randomly choose between successful closing or failed outcome

## Output Format (Strictly Follow)

### Dialogue Content
[Salesperson]: {dialogue content}
[Psychology]: {salesperson's strategic intent}
[Customer]: {dialogue content}
[Psychology]: {customer's real thoughts}
... (continue dialogue)

### Result
{Successfully Closed / Failed to Close}

### Analysis

#### Strengths
1. **{Technique Name}**
   - Quote: "{relevant dialogue segment}"
   - Analysis: {why this was effective}

#### Areas for Improvement
1. **{Issue Description}**
   - Quote: "{relevant dialogue segment}"
   - Suggestion: {specific improvement recommendation}

#### Scores (1-10)
- Opening & Ice-breaking: X
- Needs Discovery: X
- Product Introduction: X
- Objection Handling: X
- Closing: X
- Overall Performance: X

#### Key Takeaway
{Most important lesson and recommendation}

Ensure the dialogue flows naturally and the analysis is specific and constructive.`
}

// GenerateRegeneratePrompt returns a prompt for regenerating based on improvements
func GenerateRegeneratePrompt(scenario SalesScenario, improvements []string) string {
	improvementText := ""
	for i, imp := range improvements {
		improvementText += fmt.Sprintf("%d. %s\n", i+1, imp)
	}

	if scenario.Language == LanguageChinese {
		return fmt.Sprintf(`請根據以下改進建議，重新生成一段更好的銷售對話。

## 原始情境
- 產品/服務：%s
- 客人類型：%s
- 銷售員目標：%s

## 需要改進的地方
%s

## 要求
1. 針對上述每一個改進點，在新的對話中明確展現改善
2. 保持完整的銷售流程：開場、需求探索、產品介紹、處理異議、促成交易、結束
3. 對話需自然流暢，不要刻意提及「改進」
4. 重點展示「每次進步一點點」的概念
5. 請嚴格按照之前的輸出格式（對話、結果、分析、評分）

請生成改進後的完整銷售對話和分析。`, scenario.Product, scenario.CustomerType, scenario.Goal, improvementText)
	}
	return fmt.Sprintf(`Please regenerate an improved sales dialogue based on the following improvement suggestions.

## Original Scenario
- Product/Service: %s
- Customer Type: %s
- Sales Goal: %s

## Areas to Improve
%s

## Requirements
1. Address EACH improvement point clearly in the new dialogue
2. Maintain a complete sales process: opening, needs discovery, product introduction, objection handling, closing, follow-up
3. Keep the dialogue natural - do NOT mention "improvement" explicitly
4. Emphasize the concept of "small progress every day"
5. Strictly follow the output format (dialogue, result, analysis, scores)

Please generate the improved complete sales dialogue and analysis.`, scenario.Product, scenario.CustomerType, scenario.Goal, improvementText)
}

// GenerateUserPrompt returns the user-level prompt (specific scenario to generate)
func GenerateUserPrompt(scenario SalesScenario) string {
	if scenario.Language == LanguageChinese {
		return `請幫我生成一段銷售對話：

- 產品/服務：` + scenario.Product + `
- 客人類型：` + scenario.CustomerType + `（想瘦身但擔心沒時間和動力）
- 銷售員目標：` + scenario.Goal + `
- 重要提示：請在對話中多次強調「持續進步」的重要性，這是核心銷售點`
	}
	return `Please generate a sales dialogue with the following scenario:

- Product/Service: ` + scenario.Product + `
- Customer Type: ` + scenario.CustomerType + ` (wants to get fit but worried about time and motivation)
- Sales Goal: ` + scenario.Goal + `
- Important Note: Emphasize "continuous improvement" throughout the conversation - this is the key selling point`
}
