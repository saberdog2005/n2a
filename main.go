package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
	"github.com/russross/blackfriday/v2"
	"github.com/sashabaranov/go-openai"
)

// Config holds application configuration
type Config struct {
	APIKey       string  `json:"api_key"`
	Model        string  `json:"model"`
	MaxTokens    int     `json:"max_tokens"`
	Temperature  float32 `json:"temperature"`
	SystemPrompt string  `json:"system_prompt"`
}

// Flashcard represents a single Anki flashcard
type Flashcard struct {
	Front string   `json:"front"`
	Back  string   `json:"back"`
	Tags  []string `json:"tags,omitempty"`
}

// FileParser interface for different file types
type FileParser interface {
	Parse(filepath string) (string, error)
}

// PDFParser implements FileParser for PDF files
type PDFParser struct{}

func (p *PDFParser) Parse(filepath string) (string, error) {
	f, r, err := pdf.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// DOCXParser implements FileParser for DOCX files
type DOCXParser struct{}

func (d *DOCXParser) Parse(filepath string) (string, error) {
	r, err := docx.ReadDocxFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read DOCX: %w", err)
	}
	defer r.Close()

	doc := r.Editable()
	content := doc.GetContent()

	return content, nil
}

// MarkdownParser implements FileParser for Markdown files
type MarkdownParser struct{}

func (m *MarkdownParser) Parse(filepath string) (string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read markdown file: %w", err)
	}

	// Convert markdown to plain text
	html := blackfriday.Run(content)
	// Simple HTML stripping (in production, use a proper HTML parser)
	text := stripHTML(string(html))

	return text, nil
}

// stripHTML removes HTML tags from text (simplified version)
func stripHTML(html string) string {
	// This is a simplified version. In production, use golang.org/x/net/html
	var result strings.Builder
	inTag := false

	for _, r := range html {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				result.WriteRune(r)
			}
		}
	}

	return result.String()
}

// extractJSON finds and extracts JSON array from Claude's response
func extractJSON(response string) (string, error) {
	// Remove markdown code blocks first
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Find the first '[' and last ']' to extract the JSON array
	startIdx := strings.Index(response, "[")
	if startIdx == -1 {
		return "", fmt.Errorf("no JSON array found in response")
	}

	endIdx := strings.LastIndex(response, "]")
	if endIdx == -1 || endIdx < startIdx {
		return "", fmt.Errorf("malformed JSON array in response")
	}

	jsonContent := response[startIdx : endIdx+1]

	// Validate that it's valid JSON by attempting to unmarshal
	var testArray []interface{}
	if err := json.Unmarshal([]byte(jsonContent), &testArray); err != nil {
		return "", fmt.Errorf("extracted content is not valid JSON: %w", err)
	}

	return jsonContent, nil
}

// LLMClient handles interaction with the language model
type LLMClient struct {
	client *openai.Client
	config Config
}

// NewLLMClient creates a new LLM client
func NewLLMClient(config Config) *LLMClient {
	clientConfig := openai.DefaultConfig(config.APIKey)
	clientConfig.BaseURL = "https://api.anthropic.com/v1"
	client := openai.NewClientWithConfig(clientConfig)
	return &LLMClient{
		client: client,
		config: config,
	}
}

// GenerateFlashcards converts text content to flashcards
func (l *LLMClient) GenerateFlashcards(content string, subject string) ([]Flashcard, error) {
	systemPrompt := l.config.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = `You are an expert educator creating Anki flashcards. 
		Follow these principles:
		1. Create atomic cards (one concept per card)
		2. Make questions clear and unambiguous
		3. Keep answers concise but complete
		4. Focus on key concepts, definitions, formulas, and relationships
		5. Use active recall principles
		
		CRITICAL: Respond ONLY with a valid JSON array. Do not include any explanatory text, introductions, or conclusions. 
		Output format: JSON array of objects with "front" (question) and "back" (answer) fields.
		Generate comprehensive flashcards covering all important information.`
	}

	userPrompt := fmt.Sprintf(`Convert the following %s notes into Anki flashcards:

%s

Create flashcards that cover all key concepts, ensuring each card tests a single piece of knowledge.
Output as a JSON array.`, subject, content)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := l.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: l.config.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			MaxTokens:   l.config.MaxTokens,
			Temperature: l.config.Temperature,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	// Parse JSON response
	var flashcards []Flashcard
	responseContent := resp.Choices[0].Message.Content

	// Extract JSON from Claude's potentially verbose response
	jsonContent, err := extractJSON(responseContent)
	if err != nil {
		// If JSON extraction fails, show the actual response for debugging
		return nil, fmt.Errorf("failed to extract JSON from response: %w\nActual response: %s", err, responseContent)
	}

	err = json.Unmarshal([]byte(jsonContent), &flashcards)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extracted JSON: %w\nExtracted JSON: %s", err, jsonContent)
	}

	return flashcards, nil
}

// AnkiExporter handles exporting flashcards to Anki-compatible formats
type AnkiExporter struct{}

// ExportTXT exports flashcards to tab-separated text file
func (e *AnkiExporter) ExportTXT(flashcards []Flashcard, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	for _, card := range flashcards {
		// Anki format: Front[TAB]Back[TAB]Tags
		line := fmt.Sprintf("%s\t%s",
			strings.ReplaceAll(card.Front, "\t", " "),
			strings.ReplaceAll(card.Back, "\t", " "))

		if len(card.Tags) > 0 {
			line += "\t" + strings.Join(card.Tags, " ")
		}

		_, err := file.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("failed to write flashcard: %w", err)
		}
	}

	return nil
}

// ExportCSV exports flashcards to CSV format
func (e *AnkiExporter) ExportCSV(flashcards []Flashcard, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	err = writer.Write([]string{"Front", "Back", "Tags"})
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write flashcards
	for _, card := range flashcards {
		tags := strings.Join(card.Tags, " ")
		err := writer.Write([]string{card.Front, card.Back, tags})
		if err != nil {
			return fmt.Errorf("failed to write flashcard: %w", err)
		}
	}

	return nil
}

// ProcessingPipeline orchestrates the conversion process
type ProcessingPipeline struct {
	parser   FileParser
	llm      *LLMClient
	exporter *AnkiExporter
}

// NewProcessingPipeline creates a new processing pipeline
func NewProcessingPipeline(config Config) *ProcessingPipeline {
	return &ProcessingPipeline{
		llm:      NewLLMClient(config),
		exporter: &AnkiExporter{},
	}
}

// Process converts input file to Anki flashcards
func (p *ProcessingPipeline) Process(inputPath, outputPath string, dryRun bool) error {
	// Determine file type and select parser
	ext := strings.ToLower(filepath.Ext(inputPath))
	switch ext {
	case ".pdf":
		p.parser = &PDFParser{}
	case ".docx":
		p.parser = &DOCXParser{}
	case ".md", ".markdown":
		p.parser = &MarkdownParser{}
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	// Extract text content
	fmt.Println("📖 Extracting text from file...")
	content, err := p.parser.Parse(inputPath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	if len(strings.TrimSpace(content)) == 0 {
		return fmt.Errorf("no text content found in file")
	}

	fmt.Printf("✅ Extracted %d characters of text\n", len(content))

	// Determine subject from filename
	subject := strings.TrimSuffix(filepath.Base(inputPath), ext)

	// Generate flashcards using Claude AI
	fmt.Println("🤖 Generating flashcards with Claude AI...")
	flashcards, err := p.llm.GenerateFlashcards(content, subject)
	if err != nil {
		return fmt.Errorf("failed to generate flashcards: %w", err)
	}

	fmt.Printf("✅ Generated %d flashcards\n", len(flashcards))

	// Add subject as tag
	for i := range flashcards {
		flashcards[i].Tags = append(flashcards[i].Tags, subject)
	}

	// Dry run mode - preview cards
	if dryRun {
		fmt.Println("\n🔍 Preview (first 5 cards):")
		for i, card := range flashcards {
			if i >= 5 {
				break
			}
			fmt.Printf("\nCard %d:\n", i+1)
			fmt.Printf("  Front: %s\n", card.Front)
			fmt.Printf("  Back: %s\n", card.Back)
		}
		return nil
	}

	// Export to file
	fmt.Println("💾 Saving flashcards...")
	outputExt := strings.ToLower(filepath.Ext(outputPath))
	switch outputExt {
	case ".csv":
		err = p.exporter.ExportCSV(flashcards, outputPath)
	case ".txt", ".tsv":
		err = p.exporter.ExportTXT(flashcards, outputPath)
	default:
		// Default to TXT format
		err = p.exporter.ExportTXT(flashcards, outputPath)
	}

	if err != nil {
		return fmt.Errorf("failed to export flashcards: %w", err)
	}

	fmt.Printf("✅ Successfully saved %d flashcards to %s\n", len(flashcards), outputPath)
	return nil
}

// LoadConfig loads configuration from file or environment
func LoadConfig(configPath string) (Config, error) {
	config := Config{
		Model:       "claude-3-5-haiku-20241022",
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	// Load .env file if it exists
	_ = godotenv.Load()

	// Try to load from config file (if no path specified, try config.json by default)
	if configPath == "" {
		configPath = "config.json"
	}

	data, err := os.ReadFile(configPath)
	if err == nil {
		err = json.Unmarshal(data, &config)
		if err != nil {
			return config, fmt.Errorf("invalid config file: %w", err)
		}
	}

	// Override with environment variable if set
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}

	if config.APIKey == "" {
		return config, fmt.Errorf("API key not found. Set ANTHROPIC_API_KEY environment variable or provide in config file")
	}

	return config, nil
}

func main() {
	// Parse command-line arguments
	var (
		configPath string
		dryRun     bool
		help       bool
	)

	flag.StringVar(&configPath, "config", "", "Path to configuration file")
	flag.BoolVar(&dryRun, "dry-run", false, "Preview flashcards without saving")
	flag.BoolVar(&help, "help", false, "Show help message")
	flag.Parse()

	if help || flag.NArg() < 2 {
		fmt.Println("Note to Anki - Convert study notes to Anki flashcards")
		fmt.Println("\nUsage:")
		fmt.Println("  note2anki [options] <input-file> <output-file>")
		fmt.Println("\nOptions:")
		fmt.Println("  -config string   Path to configuration file")
		fmt.Println("  -dry-run        Preview flashcards without saving")
		fmt.Println("  -help           Show this help message")
		fmt.Println("\nSupported input formats: PDF, DOCX, MD")
		fmt.Println("Supported output formats: TXT (tab-separated), CSV")
		fmt.Println("\nExample:")
		fmt.Println("  note2anki notes.pdf flashcards.txt")
		fmt.Println("  note2anki -dry-run lecture.docx preview.csv")
		os.Exit(0)
	}

	inputPath := flag.Arg(0)
	outputPath := flag.Arg(1)

	// Validate input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		log.Fatalf("❌ Error: Input file does not exist: %s", inputPath)
	}

	// Load configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("❌ Configuration error: %v", err)
	}

	// Create and run pipeline
	pipeline := NewProcessingPipeline(config)

	startTime := time.Now()
	err = pipeline.Process(inputPath, outputPath, dryRun)
	if err != nil {
		log.Fatalf("❌ Processing failed: %v", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("⏱️  Completed in %.2f seconds\n", duration.Seconds())
}
