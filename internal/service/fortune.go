package service

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// FortuneServiceInterface defines the contract for our fortune service.
// Handlers will depend on this interface, not the concrete implementation.
type FortuneServiceInterface interface {
	GetFortune(opts FortuneOptions) (*FortuneResponse, error)
	ListFiles() ([]string, error)
	SearchFortunes(pattern string, opts FortuneOptions) (*SearchResponse, error)
}

// Ensure FortuneService implements the interface.
// This is a compile-time check.
var _ FortuneServiceInterface = (*FortuneService)(nil)

type FortuneService struct {
	fortunePath string
	logger      *zap.Logger
}

type FortuneOptions struct {
	All         bool     `json:"all"`
	ShowCookie  bool     `json:"show_cookie"`
	Equal       bool     `json:"equal"`
	Long        bool     `json:"long"`
	Short       bool     `json:"short"`
	IgnoreCase  bool     `json:"ignore_case"`
	Wait        bool     `json:"wait"`
	Length      int      `json:"length"`
	Pattern     string   `json:"pattern"`
	Files       []string `json:"files"`
	Percentages []string `json:"percentages"`
}

type FortuneResponse struct {
	Fortune    string `json:"fortune"`
	SourceFile string `json:"source_file,omitempty"`
}

type SearchResponse struct {
	Matches []FortuneResponse `json:"matches"`
	Count   int               `json:"count"`
}

func NewFortuneService(fortunePath string, logger *zap.Logger) *FortuneService {
	return &FortuneService{
		fortunePath: fortunePath,
		logger:      logger,
	}
}

func (s *FortuneService) GetFortune(opts FortuneOptions) (*FortuneResponse, error) {
	args := s.buildArgs(opts)

	cmd := exec.Command(s.fortunePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("Fortune command failed",
			zap.Error(err),
			zap.String("command", s.fortunePath),
			zap.Strings("args", args),
			zap.String("output", string(output)))
		return nil, fmt.Errorf("fortune command failed: %w", err)
	}

	fortune := strings.TrimSpace(string(output))
	if fortune == "" {
		return nil, errors.New("no fortune returned")
	}

	response := &FortuneResponse{
		Fortune: fortune,
	}

	// If show_cookie is enabled, we might have source file info
	if opts.ShowCookie {
		lines := strings.Split(fortune, "\n")
		if len(lines) > 1 && strings.Contains(lines[len(lines)-1], "(") {
			response.SourceFile = strings.Trim(lines[len(lines)-1], "()")
			response.Fortune = strings.Join(lines[:len(lines)-1], "\n")
		}
	}

	return response, nil
}

func (s *FortuneService) ListFiles() ([]string, error) {
	// On Debian systems, fortune files are stored in this directory.
	const fortuneDir = "/usr/share/games/fortunes/"

	entries, err := os.ReadDir(fortuneDir)
	if err != nil {
		s.logger.Error("Failed to read fortune directory", zap.Error(err), zap.String("directory", fortuneDir))
		return nil, fmt.Errorf("could not read fortune directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		// We only want the actual data files, not directories or the index (.dat) files.
		if !entry.IsDir() && !strings.HasSuffix(entry.Name(), ".dat") {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

func (s *FortuneService) SearchFortunes(pattern string, opts FortuneOptions) (*SearchResponse, error) {
	if pattern == "" {
		return nil, errors.New("search pattern is required")
	}

	// Force pattern search
	opts.Pattern = pattern
	args := s.buildArgs(opts)

	cmd := exec.Command(s.fortunePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("Fortune search command failed", zap.Error(err))
		return nil, fmt.Errorf("fortune search failed: %w", err)
	}

	matches := s.parseSearchResults(string(output))

	return &SearchResponse{
		Matches: matches,
		Count:   len(matches),
	}, nil
}

func (s *FortuneService) buildArgs(opts FortuneOptions) []string {
	var args []string

	if opts.All {
		args = append(args, "-a")
	}
	if opts.ShowCookie {
		args = append(args, "-c")
	}
	if opts.Equal {
		args = append(args, "-e")
	}
	if opts.Long {
		args = append(args, "-l")
	}
	if opts.Short {
		args = append(args, "-s")
	}
	if opts.IgnoreCase {
		args = append(args, "-i")
	}
	if opts.Wait {
		args = append(args, "-w")
	}
	if opts.Length > 0 {
		args = append(args, "-n", strconv.Itoa(opts.Length))
	}
	if opts.Pattern != "" {
		args = append(args, "-m", opts.Pattern)
	}

	// Add file specifications with percentages
	for i, file := range opts.Files {
		if i < len(opts.Percentages) && opts.Percentages[i] != "" {
			args = append(args, opts.Percentages[i]+"%", file)
		} else {
			args = append(args, file)
		}
	}

	return args
}

// This function is now corrected to use the proper separator.
func (s *FortuneService) parseSearchResults(output string) []FortuneResponse {
	var matches []FortuneResponse

	// The `fortune -m` command separates matches with a '%' on its own line.
	fortunes := strings.Split(output, "\n%\n")

	for _, fortune := range fortunes {
		fortune = strings.TrimSpace(fortune)
		if fortune != "" {
			matches = append(matches, FortuneResponse{
				Fortune: fortune,
			})
		}
	}

	return matches
}
