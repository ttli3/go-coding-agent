// formats terminal output 
package ui

import (
	"bufio"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

type ResponseFormatter struct {
	// Color definitions
	headerColor    *color.Color
	codeColor      *color.Color
	commentColor   *color.Color
	stringColor    *color.Color
	keywordColor   *color.Color
	numberColor    *color.Color
	operatorColor  *color.Color
	bulletColor    *color.Color
	emphasisColor  *color.Color
	linkColor      *color.Color
	quoteColor     *color.Color
	separatorColor *color.Color
}

func NewResponseFormatter() *ResponseFormatter {
	return &ResponseFormatter{
		headerColor:    color.New(color.FgCyan, color.Bold),
		codeColor:      color.New(color.FgYellow),
		commentColor:   color.New(color.FgHiBlack),
		stringColor:    color.New(color.FgGreen),
		keywordColor:   color.New(color.FgMagenta, color.Bold),
		numberColor:    color.New(color.FgBlue),
		operatorColor:  color.New(color.FgRed),
		bulletColor:    color.New(color.FgCyan),
		emphasisColor:  color.New(color.FgWhite, color.Bold),
		linkColor:      color.New(color.FgBlue, color.Underline),
		quoteColor:     color.New(color.FgHiBlack, color.Italic),
		separatorColor: color.New(color.FgHiBlack),
	}
}

func (f *ResponseFormatter) FormatResponse(response string) string {
	var result strings.Builder

	// Split response into lines for processing
	scanner := bufio.NewScanner(strings.NewReader(response))
	inCodeBlock := false
	codeLanguage := ""

	for scanner.Scan() {
		line := scanner.Text()

		// Handle code blocks
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				inCodeBlock = false
				codeLanguage = ""
				result.WriteString(f.separatorColor.Sprint("└" + strings.Repeat("─", 60) + "┘\n"))
			} else {
				inCodeBlock = true
				if len(line) > 3 {
					codeLanguage = strings.TrimSpace(line[3:])
				}
				langLabel := codeLanguage
				if langLabel == "" {
					langLabel = "code"
				}
				result.WriteString(f.separatorColor.Sprint("┌─ "))
				result.WriteString(f.codeColor.Sprint(langLabel))
				result.WriteString(f.separatorColor.Sprint(" " + strings.Repeat("─", 55-len(langLabel)) + "┐\n"))
			}
			continue
		}

		if inCodeBlock {
			// Format code inside code blocks
			result.WriteString(f.separatorColor.Sprint("│ "))
			result.WriteString(f.formatCodeLine(line, codeLanguage))
			result.WriteString("\n")
		} else {
			// Format regular markdown content
			result.WriteString(f.formatMarkdownLine(line))
			result.WriteString("\n")
		}
	}

	return result.String()
}

func (f *ResponseFormatter) formatCodeLine(line, language string) string {
	// Basic syntax highlighting for common languages
	switch strings.ToLower(language) {
	case "go", "golang":
		return f.formatGoCode(line)
	case "javascript", "js":
		return f.formatJavaScriptCode(line)
	case "python", "py":
		return f.formatPythonCode(line)
	case "bash", "shell", "sh":
		return f.formatBashCode(line)
	default:
		return f.formatGenericCode(line)
	}
}

func (f *ResponseFormatter) formatGoCode(line string) string {
	// Go keywords
	goKeywords := []string{
		"package", "import", "func", "var", "const", "type", "struct", "interface",
		"if", "else", "for", "range", "switch", "case", "default", "break", "continue",
		"return", "go", "defer", "chan", "select", "map", "make", "new", "len", "cap",
	}

	result := line

	// Highlight keywords
	for _, keyword := range goKeywords {
		pattern := `\b` + regexp.QuoteMeta(keyword) + `\b`
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			return f.keywordColor.Sprint(match)
		})
	}

	// Highlight strings
	stringRe := regexp.MustCompile(`"[^"]*"`)
	result = stringRe.ReplaceAllStringFunc(result, func(match string) string {
		return f.stringColor.Sprint(match)
	})

	// Highlight comments
	commentRe := regexp.MustCompile(`//.*$`)
	result = commentRe.ReplaceAllStringFunc(result, func(match string) string {
		return f.commentColor.Sprint(match)
	})

	// Highlight numbers
	numberRe := regexp.MustCompile(`\b\d+\b`)
	result = numberRe.ReplaceAllStringFunc(result, func(match string) string {
		return f.numberColor.Sprint(match)
	})

	return result
}

func (f *ResponseFormatter) formatJavaScriptCode(line string) string {
	jsKeywords := []string{
		"function", "var", "let", "const", "if", "else", "for", "while", "do",
		"switch", "case", "default", "break", "continue", "return", "try", "catch",
		"finally", "throw", "new", "this", "typeof", "instanceof", "in", "of",
	}

	result := line

	for _, keyword := range jsKeywords {
		pattern := `\b` + regexp.QuoteMeta(keyword) + `\b`
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			return f.keywordColor.Sprint(match)
		})
	}

	// Strings
	stringRe := regexp.MustCompile(`["'][^"']*["']`)
	result = stringRe.ReplaceAllStringFunc(result, func(match string) string {
		return f.stringColor.Sprint(match)
	})

	// Comments
	commentRe := regexp.MustCompile(`//.*$`)
	result = commentRe.ReplaceAllStringFunc(result, func(match string) string {
		return f.commentColor.Sprint(match)
	})

	return result
}

func (f *ResponseFormatter) formatPythonCode(line string) string {
	pythonKeywords := []string{
		"def", "class", "if", "elif", "else", "for", "while", "try", "except",
		"finally", "with", "as", "import", "from", "return", "yield", "break",
		"continue", "pass", "raise", "assert", "del", "global", "nonlocal",
		"lambda", "and", "or", "not", "in", "is", "True", "False", "None",
	}

	result := line

	for _, keyword := range pythonKeywords {
		pattern := `\b` + regexp.QuoteMeta(keyword) + `\b`
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			return f.keywordColor.Sprint(match)
		})
	}

	// Strings
	stringRe := regexp.MustCompile(`["'][^"']*["']`)
	result = stringRe.ReplaceAllStringFunc(result, func(match string) string {
		return f.stringColor.Sprint(match)
	})

	// Comments
	commentRe := regexp.MustCompile(`#.*$`)
	result = commentRe.ReplaceAllStringFunc(result, func(match string) string {
		return f.commentColor.Sprint(match)
	})

	return result
}

func (f *ResponseFormatter) formatBashCode(line string) string {
	bashKeywords := []string{
		"if", "then", "else", "elif", "fi", "for", "while", "do", "done",
		"case", "esac", "function", "return", "exit", "break", "continue",
		"echo", "printf", "read", "cd", "ls", "mkdir", "rm", "cp", "mv",
	}

	result := line

	for _, keyword := range bashKeywords {
		pattern := `\b` + regexp.QuoteMeta(keyword) + `\b`
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			return f.keywordColor.Sprint(match)
		})
	}

	// Strings
	stringRe := regexp.MustCompile(`["'][^"']*["']`)
	result = stringRe.ReplaceAllStringFunc(result, func(match string) string {
		return f.stringColor.Sprint(match)
	})

	// Comments
	commentRe := regexp.MustCompile(`#.*$`)
	result = commentRe.ReplaceAllStringFunc(result, func(match string) string {
		return f.commentColor.Sprint(match)
	})

	// Variables
	varRe := regexp.MustCompile(`\$\w+`)
	result = varRe.ReplaceAllStringFunc(result, func(match string) string {
		return f.numberColor.Sprint(match)
	})

	return result
}

func (f *ResponseFormatter) formatGenericCode(line string) string {
	result := line

	// Highlight strings
	stringRe := regexp.MustCompile(`["'][^"']*["']`)
	result = stringRe.ReplaceAllStringFunc(result, func(match string) string {
		return f.stringColor.Sprint(match)
	})

	// Highlight numbers
	numberRe := regexp.MustCompile(`\b\d+\b`)
	result = numberRe.ReplaceAllStringFunc(result, func(match string) string {
		return f.numberColor.Sprint(match)
	})

	return result
}

func (f *ResponseFormatter) formatMarkdownLine(line string) string {
	// Handle headers
	if strings.HasPrefix(line, "#") {
		level := 0
		for i, char := range line {
			if char == '#' {
				level++
			} else if char == ' ' {
				headerText := line[i+1:]
				return f.formatHeader(headerText, level)
			} else {
				break
			}
		}
	}

	// Handle bullet points
	if strings.HasPrefix(strings.TrimSpace(line), "- ") || strings.HasPrefix(strings.TrimSpace(line), "* ") {
		return f.formatBulletPoint(line)
	}

	// Handle numbered lists
	numberedRe := regexp.MustCompile(`^\s*\d+\.\s+`)
	if numberedRe.MatchString(line) {
		return f.formatNumberedList(line)
	}

	// Handle quotes
	if strings.HasPrefix(strings.TrimSpace(line), "> ") {
		return f.formatQuote(line)
	}

	// Handle inline formatting
	return f.formatInlineElements(line)
}

func (f *ResponseFormatter) formatHeader(text string, level int) string {
	var prefix string
	switch level {
	case 1:
		prefix = "█ "
	case 2:
		prefix = "▓ "
	case 3:
		prefix = "▒ "
	default:
		prefix = "░ "
	}

	return f.headerColor.Sprint(prefix + text)
}

func (f *ResponseFormatter) formatBulletPoint(line string) string {
	trimmed := strings.TrimSpace(line)
	content := trimmed[2:] // Remove "- " or "* "
	indent := len(line) - len(trimmed)

	return strings.Repeat(" ", indent) + f.bulletColor.Sprint("• ") + f.formatInlineElements(content)
}

func (f *ResponseFormatter) formatNumberedList(line string) string {
	re := regexp.MustCompile(`^(\s*)(\d+\.)(\s+)(.*)$`)
	matches := re.FindStringSubmatch(line)
	if len(matches) == 5 {
		indent := matches[1]
		number := matches[2]
		space := matches[3]
		content := matches[4]

		return indent + f.bulletColor.Sprint(number) + space + f.formatInlineElements(content)
	}
	return line
}

func (f *ResponseFormatter) formatQuote(line string) string {
	trimmed := strings.TrimSpace(line)
	content := trimmed[2:] // Remove "> "
	indent := len(line) - len(trimmed)

	return strings.Repeat(" ", indent) + f.quoteColor.Sprint("▌ "+content)
}

func (f *ResponseFormatter) formatInlineElements(text string) string {
	result := text

	// Bold text **text**
	boldRe := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	result = boldRe.ReplaceAllStringFunc(result, func(match string) string {
		content := match[2 : len(match)-2] // Remove ** from both ends
		return f.emphasisColor.Sprint(content)
	})

	// Italic text *text*
	italicRe := regexp.MustCompile(`\*([^*]+)\*`)
	result = italicRe.ReplaceAllStringFunc(result, func(match string) string {
		content := match[1 : len(match)-1] // Remove * from both ends
		return color.New(color.Italic).Sprint(content)
	})

	// Inline code `code`
	codeRe := regexp.MustCompile("`([^`]+)`")
	result = codeRe.ReplaceAllStringFunc(result, func(match string) string {
		content := match[1 : len(match)-1] // Remove ` from both ends
		return f.codeColor.Sprint(content)
	})

	// Links [text](url)
	linkRe := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	result = linkRe.ReplaceAllStringFunc(result, func(match string) string {
		matches := linkRe.FindStringSubmatch(match)
		if len(matches) == 3 {
			text := matches[1]
			url := matches[2]
			return f.linkColor.Sprint(text) + f.commentColor.Sprint(" ("+url+")")
		}
		return match
	})

	return result
}

// FormatStreamingChunk formats individual chunks for streaming responses
func (f *ResponseFormatter) FormatStreamingChunk(chunk string) string {
	// For streaming, we'll do minimal formatting to avoid breaking mid-word
	// Just handle basic inline code and emphasis
	result := chunk

	// Only format complete inline elements
	codeRe := regexp.MustCompile("`([^`]+)`")
	result = codeRe.ReplaceAllStringFunc(result, func(match string) string {
		content := match[1 : len(match)-1]
		return f.codeColor.Sprint(content)
	})

	return result
}
