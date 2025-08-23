# Note2Anki - Study Notes to Anki Flashcards Converter

A powerful CLI tool that converts your study notes (PDF, DOCX, Markdown) into Anki-compatible flashcards using Claude AI.

## Features

- 📄 **Multiple Input Formats**: Supports PDF, DOCX, and Markdown files
- 🤖 **AI-Powered**: Uses Anthropic's Claude models to intelligently generate flashcards
- 📝 **Anki-Compatible**: Exports to TXT (tab-separated) and CSV formats
- 🎯 **Smart Card Generation**: Creates atomic, focused flashcards following spaced repetition best practices
- 🏷️ **Auto-Tagging**: Automatically tags cards based on the source file
- 👀 **Preview Mode**: Dry-run option to preview cards before saving
- ⚡ **Fast Processing**: Typically processes notes in under 30 seconds
- 🛠️ **Configurable**: Customizable prompts and LLM parameters

## Installation

### Prerequisites

- Go 1.21 or higher
- Anthropic API key

### Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/note2anki.git
cd note2anki

# Install dependencies
go mod download

# Build the binary
go build -o note2anki main.go

# Optional: Install globally
go install
```

## Configuration

### Method 1: .env File (Recommended)

Create a `.env` file in your project root:

```bash
ANTHROPIC_API_KEY=your-anthropic-api-key-here
```

### Method 2: Environment Variable

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

### Method 3: Configuration File

Create a `config.json` file:

```json
{
  "api_key": "your-anthropic-api-key",
  "model": "claude-3-5-haiku-20241022",
  "max_tokens": 2000,
  "temperature": 0.7
}
```

## 8. After making these changes, run:

```bash
go mod tidy
go build -o build/note2anki main.go
```

## Usage

### Basic Usage

```bash
note2anki <input-file> <output-file>
```

### Examples

```bash
# Convert PDF notes to Anki flashcards
note2anki biology_notes.pdf biology_cards.txt

# Convert DOCX to CSV format
note2anki lecture.docx flashcards.csv

# Convert Markdown notes
note2anki chemistry.md chemistry_cards.txt

# Preview cards without saving (dry-run)
note2anki -dry-run physics.pdf preview.txt

# Use custom configuration
note2anki -config custom_config.json notes.pdf cards.txt
```

### Command-Line Options

- `-config <path>`: Path to configuration file
- `-dry-run`: Preview first 5 flashcards without saving
- `-help`: Show help message

## Output Formats

### Tab-Separated Text (.txt)

Default Anki import format:
```
Question[TAB]Answer[TAB]Tags
What is photosynthesis?[TAB]The process by which plants convert light energy into chemical energy[TAB]biology
```

### CSV Format (.csv)

Spreadsheet-compatible format with headers:
```csv
Front,Back,Tags
"What is photosynthesis?","The process by which plants convert light energy into chemical energy","biology"
```

## How It Works

1. **Text Extraction**: Parses your input file and extracts all text content
2. **AI Processing**: Sends content to Anthropic's Claude model with specialized prompts
3. **Card Generation**: AI identifies key concepts and creates question-answer pairs
4. **Quality Control**: Ensures cards follow Anki best practices (atomic, clear, testable)
5. **Export**: Saves flashcards in Anki-compatible format

## Flashcard Quality Principles

The tool generates flashcards following these principles:

- **Atomic**: One concept per card for optimal retention
- **Clear**: Unambiguous questions with precise answers
- **Contextual**: Preserves important context from notes
- **Comprehensive**: Covers all key concepts, definitions, and relationships
- **Active Recall**: Questions designed to promote active memory retrieval

## Advanced Configuration

### Custom System Prompts

You can customize the AI's behavior by modifying the system prompt in your config file:

```json
{
  "system_prompt": "Create medical school flashcards focusing on clinical applications..."
}
```

### Model Selection

Choose different Claude models based on your needs:

- `claude-3-5-haiku-20241022`: Fast and cost-effective (default)
- `claude-3-5-sonnet-20241022`: More accurate and capable
- `claude-3-opus-20240229`: Most powerful for complex tasks

## Troubleshooting

### Common Issues

1. **"API key not found"**
   - Set the `ANTHROPIC_API_KEY` environment variable
   - Or provide it in a config file

2. **"No text content found"**
   - Ensure the PDF is not image-based (scanned documents)
   - Check that the file is not corrupted

3. **"LLM request failed"**
   - Verify your API key is valid
   - Check your Anthropic account has credits
   - Ensure you have internet connectivity

4. **Large files taking too long**
   - Consider splitting very large documents
   - Adjust `max_tokens` in configuration

## Best Practices

1. **Organize Your Notes**: Well-structured notes produce better flashcards
2. **Use Clear Headings**: Help the AI understand topic boundaries
3. **Include Examples**: Examples in notes often become great flashcard content
4. **Review Generated Cards**: Always review Claude-generated content before studying
5. **Batch Processing**: Process related notes together for consistent tagging

## Performance

- Small files (<10 pages): ~5-10 seconds
- Medium files (10-50 pages): ~15-20 seconds
- Large files (50+ pages): ~20-30 seconds

*Note: Processing time depends on file complexity and API response time*

## Importing to Anki

1. Open Anki Desktop
2. Click "File" → "Import"
3. Select your generated file
4. Configure import settings:
   - For TXT: Set field separator to "Tab"
   - For CSV: Anki will auto-detect
5. Map fields to your note type
6. Click "Import"

## Privacy & Security

- Your notes are sent to Anthropic for processing
- No data is stored by this tool
- API keys are never logged or transmitted except to Anthropic
- Consider sensitive content before processing

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

MIT License - See LICENSE file for details

## Roadmap

- [ ] Support for more file formats (RTF, HTML, EPUB)
- [ ] Batch processing multiple files
- [ ] Local LLM support (Ollama integration)
- [ ] Direct .apkg file generation
- [ ] Web interface option
- [ ] Custom note types support
- [ ] Image-based flashcard support
- [ ] Cloze deletion cards

## Support

For issues, questions, or suggestions, please open an issue on GitHub.