package judge

// LangConfig defines how a programming language is compiled and executed.
//
// WHY a config struct instead of if-else chains?
// Because adding a new language becomes a ONE-LINE change — just add a new
// entry to the SupportedLanguages map. No logic changes needed.
//
// Interview answer: "I use a config-driven approach. Each language is a
// struct with its Docker image, compile command, run command, and file
// extension. Adding Go support is literally adding one map entry."
type LangConfig struct {
	// Docker image to use (e.g., "judge-sandbox-cpp", "judge-sandbox-python")
	Image string

	// File extension for the source code (e.g., ".cpp", ".py", ".java")
	FileExtension string

	// CompileCmd is the shell command to compile the code.
	// Empty string means the language is interpreted (no compilation step).
	// Placeholders: {source} = source file path, {output} = binary output path
	CompileCmd string

	// RunCmd is the command to execute the program.
	// For compiled languages: path to the binary
	// For interpreted languages: interpreter + source file
	RunCmd string

	// NeedsCompilation is true for compiled languages (C++, Java, Go)
	// and false for interpreted languages (Python, JavaScript)
	NeedsCompilation bool
}

// SupportedLanguages maps language name → config.
// The frontend sends the language name (e.g., "cpp", "python", "java")
// and the worker looks up the config here.
//
// To add a new language:
//   1. Create a Dockerfile in server/docker/
//   2. Build it: docker build -t judge-sandbox-<lang> -f Dockerfile.<lang> .
//   3. Add an entry here
var SupportedLanguages = map[string]LangConfig{
	"cpp": {
		Image:            "judge-sandbox",
		FileExtension:    ".cpp",
		CompileCmd:       "g++ /sandbox/code.cpp -o /sandbox/main -O2 -std=c++17 -Wall",
		RunCmd:           "/sandbox/main",
		NeedsCompilation: true,
	},
	"python": {
		Image:            "judge-sandbox-python",
		FileExtension:    ".py",
		CompileCmd:       "", // Python is interpreted — no compilation
		RunCmd:           "python3 /sandbox/code.py",
		NeedsCompilation: false,
	},
	"java": {
		Image:            "judge-sandbox-java",
		FileExtension:    ".java",
		CompileCmd:       "javac /sandbox/Main.java",
		RunCmd:           "java -cp /sandbox Main",
		NeedsCompilation: true,
	},
}

// GetLangConfig returns the config for a language, defaulting to C++.
func GetLangConfig(language string) LangConfig {
	if config, ok := SupportedLanguages[language]; ok {
		return config
	}
	// Default to C++ for backward compatibility
	return SupportedLanguages["cpp"]
}
