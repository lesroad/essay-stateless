package evaluate

import (
	"regexp"
	"strings"
)

// ContentCleaner 内容清理器
type ContentCleaner struct{}

// NewContentCleaner 创建内容清理器
func NewContentCleaner() *ContentCleaner {
	return &ContentCleaner{}
}

// Clean 清理作文内容中的多余换行符和特殊符号
func (c *ContentCleaner) Clean(content string) string {
	if content == "" {
		return content
	}

	// 1. 将连续的多个\n替换为单个\n
	re := regexp.MustCompile(`\n+`)
	content = re.ReplaceAllString(content, "\n")

	// 2. 去除开头和结尾的换行符
	content = strings.Trim(content, "\n")

	// 3. 清理非正常作文标点的特殊符号
	// 保留中文字符、英文字母、数字、空白字符、中英文标点
	validChars := regexp.MustCompile(`[^\p{Han}a-zA-Z0-9\s。，！？；：""''（）【】《》、.,!?;:"'()\-]`)
	content = validChars.ReplaceAllString(content, "")

	// 4. 清理多余的空格
	spaceRe := regexp.MustCompile(`[ \t]+`)
	content = spaceRe.ReplaceAllString(content, " ")

	// 5. 去除行首行尾的空格
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	content = strings.Join(lines, "\n")

	// 6. 再次去除开头和结尾的换行符
	content = strings.Trim(content, "\n")

	return content
}

