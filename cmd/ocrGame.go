package cmd

//import (
//	"bufio"
//	"bytes"
//	"encoding/json"
//	"errors"
//	"fmt"
//	"io"
//	"os"
//	"os/exec"
//	"path/filepath"
//	"sort"
//	"strconv"
//	"strings"
//	"time"
//
//
//	"github.com/spf13/cobra"
//)
//
//type processMatch struct {
//	PID  int
//	Args string
//}
//
//type hyprClient struct {
//	Address      string `json:"address"`
//	Class        string `json:"class"`
//	InitialClass string `json:"initialClass"`
//	Title        string `json:"title"`
//	InitialTitle string `json:"initialTitle"`
//	PID          int    `json:"pid"`
//	At           []int  `json:"at"`
//	Size         []int  `json:"size"`
//}
//
//type windowMatch struct {
//	Client hyprClient
//	Score  int
//	Reason string
//}
//
//var ocrGameName string
//var ocrGameLang string
//var ocrGamePSM int
//var ocrGameFormat string
//var ocrGameNoFocus bool
//var ocrGameWindowOnly bool
//var ocrGameReplaceOutput bool
//
//type ocrOutputRenderer struct {
//	writer  io.Writer
//	replace bool
//}
//
//var ocrGameCmd = &cobra.Command{
//	Use:     "ocr-game [game-name]",
//	Aliases: []string{"ocrGame"},
//	Short:   "Capture Japanese text from a running game with OCR on Hyprland",
//	Args:    cobra.MaximumNArgs(1),
//	RunE: func(cmd *cobra.Command, args []string) error {
//		selectedName := strings.TrimSpace(ocrGameName)
//		if selectedName == "" && len(args) > 0 {
//			selectedName = args[0]
//		}
//
//		var cfg *gameconfig.GameConfig
//		var err error
//		if selectedName == "" {
//
//			picker, err := launcher.NewPicker("Select a game to run OCR with", "ocr")
//			if err != nil {
//				return err
//			}
//			cfg, err = picker.SelectGameConfig()
//			if err != nil {
//				return err
//			}
//		} else {
//			cfg, err = gameconfig.FindConfig(selectedName)
//			if err != nil {
//				return err
//			}
//		}
//
//		if err := ensureOCRDependencies(); err != nil {
//			return err
//		}
//
//		match, err := findRunningGameWindow(*cfg)
//		if err != nil {
//			return err
//		}
//
//		fmt.Fprintf(cmd.OutOrStdout(), "selected game: %s\n", cfg.Name)
//		fmt.Fprintf(cmd.OutOrStdout(), "matched window pid=%d title=%q class=%q (%s)\n", match.Client.PID, match.Client.Title, match.Client.Class, match.Reason)
//		fmt.Fprintf(cmd.OutOrStdout(), "ocr output format: %s\n", ocrGameFormat)
//		fmt.Fprintln(cmd.OutOrStdout(), "draw a capture box with the mouse when prompted")
//		fmt.Fprintln(cmd.OutOrStdout(), "after each OCR result: press Enter to capture again, or type q to quit")
//
//		renderer := ocrOutputRenderer{
//			writer:  cmd.OutOrStdout(),
//			replace: ocrGameReplaceOutput,
//		}
//		return runOCRLoop(cmd, match, renderer)
//	},
//}
//
//func init() {
//	rootCmd.AddCommand(ocrGameCmd)
//	ocrGameCmd.Flags().StringVarP(&ocrGameName, "game", "g", "", "name of the saved game to OCR")
//	ocrGameCmd.Flags().StringVar(&ocrGameLang, "lang", "jpn", "tesseract OCR language")
//	ocrGameCmd.Flags().IntVar(&ocrGamePSM, "psm", 6, "tesseract page segmentation mode")
//	ocrGameCmd.Flags().StringVar(&ocrGameFormat, "format", "text", "tesseract output format: text, tsv, or hocr")
//	ocrGameCmd.Flags().BoolVar(&ocrGameNoFocus, "no-focus", false, "do not focus the matched window before each capture")
//	ocrGameCmd.Flags().BoolVar(&ocrGameWindowOnly, "window-only", false, "capture the full matched game window instead of selecting a region each time")
//	ocrGameCmd.Flags().BoolVar(&ocrGameReplaceOutput, "replace-output", false, "replace the previous OCR block on screen instead of appending to terminal scrollback")
//}
//
//func ensureOCRDependencies() error {
//	required := []string{"hyprctl", "grim", "slurp", "tesseract"}
//	var missing []string
//	for _, tool := range required {
//		if _, err := exec.LookPath(tool); err != nil {
//			missing = append(missing, tool)
//		}
//	}
//	if len(missing) > 0 {
//		return fmt.Errorf("missing required tools: %s", strings.Join(missing, ", "))
//	}
//
//	if strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) == "" {
//		return errors.New("WAYLAND_DISPLAY is not set; this OCR flow is intended for Wayland/Hyprland")
//	}
//
//	if err := ensureTesseractLanguage(ocrGameLang); err != nil {
//		return err
//	}
//	if _, err := normalizedOCRFormat(); err != nil {
//		return err
//	}
//	return nil
//}
//
//func normalizedOCRFormat() (string, error) {
//	format := strings.ToLower(strings.TrimSpace(ocrGameFormat))
//	switch format {
//	case "", "text", "txt":
//		return "text", nil
//	case "tsv":
//		return "tsv", nil
//	case "hocr":
//		return "hocr", nil
//	default:
//		return "", fmt.Errorf("unsupported OCR format %q; expected text, tsv, or hocr", ocrGameFormat)
//	}
//}
//
//func ensureTesseractLanguage(lang string) error {
//	cmd := exec.Command("tesseract", "--list-langs")
//	output, err := cmd.CombinedOutput()
//	if err != nil {
//		return fmt.Errorf("list tesseract languages: %w", err)
//	}
//	for _, line := range strings.Split(string(output), "\n") {
//		if strings.TrimSpace(line) == lang {
//			return nil
//		}
//	}
//	return fmt.Errorf("tesseract language %q is not installed", lang)
//}
//
//func findRunningGameWindow(cfg gameconfig.GameConfig) (windowMatch, error) {
//	processes, err := listProcesses()
//	if err != nil {
//		return windowMatch{}, err
//	}
//
//	clients, err := listHyprClients()
//	if err != nil {
//		return windowMatch{}, err
//	}
//	if len(clients) == 0 {
//		return windowMatch{}, errors.New("hyprctl returned no windows")
//	}
//
//	processCandidates := rankProcessMatches(cfg, processes)
//	clientMatch, ok := rankWindowMatches(cfg, clients, processCandidates)
//	if !ok {
//		return windowMatch{}, fmt.Errorf("could not find a running Hyprland window for %q", cfg.Name)
//	}
//	return clientMatch, nil
//}
//
//func listProcesses() ([]processMatch, error) {
//	cmd := exec.Command("ps", "-eo", "pid=,args=")
//	output, err := cmd.Output()
//	if err != nil {
//		return nil, fmt.Errorf("list processes: %w", err)
//	}
//
//	var processes []processMatch
//	for _, line := range strings.Split(string(output), "\n") {
//		line = strings.TrimSpace(line)
//		if line == "" {
//			continue
//		}
//
//		fields := strings.Fields(line)
//		if len(fields) < 2 {
//			continue
//		}
//
//		pid, err := strconv.Atoi(fields[0])
//		if err != nil {
//			continue
//		}
//
//		args := strings.TrimSpace(strings.TrimPrefix(line, fields[0]))
//		processes = append(processes, processMatch{
//			PID:  pid,
//			Args: args,
//		})
//	}
//	return processes, nil
//}
//
//func listHyprClients() ([]hyprClient, error) {
//	cmd := exec.Command("hyprctl", "clients", "-j")
//	output, err := cmd.Output()
//	if err != nil {
//		return nil, fmt.Errorf("query hypr clients: %w", err)
//	}
//
//	var clients []hyprClient
//	if err := json.Unmarshal(output, &clients); err != nil {
//		return nil, fmt.Errorf("decode hypr clients: %w", err)
//	}
//	return clients, nil
//}
//
//func rankProcessMatches(cfg gameconfig.GameConfig, processes []processMatch) []processMatch {
//	type scoredProcess struct {
//		process processMatch
//		score   int
//	}
//
//	executableBase := normalizedName(filepath.Base(cfg.Executable))
//	gameName := normalizedName(cfg.Name)
//	steamAppID := strings.TrimSpace(cfg.SteamAppID)
//
//	var scored []scoredProcess
//	for _, process := range processes {
//		args := strings.ToLower(process.Args)
//		score := 0
//
//		if executableBase != "" && strings.Contains(normalizedName(args), executableBase) {
//			score += 100
//		}
//		if gameName != "" && strings.Contains(normalizedName(args), gameName) {
//			score += 60
//		}
//		if steamAppID != "" && strings.Contains(args, steamAppID) {
//			score += 50
//		}
//		if score == 0 {
//			continue
//		}
//
//		scored = append(scored, scoredProcess{process: process, score: score})
//	}
//
//	sort.Slice(scored, func(i, j int) bool {
//		if scored[i].score == scored[j].score {
//			return scored[i].process.PID < scored[j].process.PID
//		}
//		return scored[i].score > scored[j].score
//	})
//
//	matches := make([]processMatch, 0, len(scored))
//	for _, item := range scored {
//		matches = append(matches, item.process)
//	}
//	return matches
//}
//
//func rankWindowMatches(cfg gameconfig.GameConfig, clients []hyprClient, processCandidates []processMatch) (windowMatch, bool) {
//	pidScores := make(map[int]int, len(processCandidates))
//	for idx, process := range processCandidates {
//		pidScores[process.PID] = 1000 - idx*10
//	}
//
//	gameName := normalizedName(cfg.Name)
//	executableBase := normalizedName(filepath.Base(cfg.Executable))
//	steamAppID := strings.TrimSpace(cfg.SteamAppID)
//
//	var best windowMatch
//	bestFound := false
//
//	for _, client := range clients {
//		score := 0
//		reasons := make([]string, 0, 3)
//		if pidScore, ok := pidScores[client.PID]; ok {
//			score += pidScore
//			reasons = append(reasons, "pid match")
//		}
//
//		textFields := []string{
//			client.Class,
//			client.InitialClass,
//			client.Title,
//			client.InitialTitle,
//		}
//		for _, field := range textFields {
//			normalizedField := normalizedName(field)
//			lowerField := strings.ToLower(field)
//			if gameName != "" && strings.Contains(normalizedField, gameName) {
//				score += 80
//				reasons = append(reasons, "name match")
//				break
//			}
//			if executableBase != "" && strings.Contains(normalizedField, executableBase) {
//				score += 70
//				reasons = append(reasons, "executable match")
//				break
//			}
//			if steamAppID != "" && strings.Contains(lowerField, steamAppID) {
//				score += 40
//				reasons = append(reasons, "steam app id match")
//				break
//			}
//		}
//
//		if score == 0 {
//			continue
//		}
//
//		match := windowMatch{
//			Client: client,
//			Score:  score,
//			Reason: strings.Join(reasons, ", "),
//		}
//		if !bestFound || match.Score > best.Score {
//			best = match
//			bestFound = true
//		}
//	}
//
//	return best, bestFound
//}
//
//func normalizedName(input string) string {
//	input = strings.ToLower(strings.TrimSpace(input))
//	var builder strings.Builder
//	builder.Grow(len(input))
//	for _, r := range input {
//		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
//			builder.WriteRune(r)
//		}
//	}
//	return builder.String()
//}
//
//func runOCRLoop(cmd *cobra.Command, match windowMatch, renderer ocrOutputRenderer) error {
//	reader := bufio.NewReader(os.Stdin)
//	tempDir, err := os.MkdirTemp("", "wgl-ocr-*")
//	if err != nil {
//		return fmt.Errorf("create temp directory: %w", err)
//	}
//	defer os.RemoveAll(tempDir)
//
//	captureNumber := 1
//	for {
//		if err := captureAndPrintOCR(match, tempDir, renderer, captureNumber); err != nil {
//			return err
//		}
//		captureNumber++
//
//		fmt.Fprint(cmd.OutOrStdout(), "\n[capture again: Enter, quit: q] ")
//		input, err := reader.ReadString('\n')
//		if err != nil && !errors.Is(err, os.ErrClosed) {
//			return fmt.Errorf("read input: %w", err)
//		}
//
//		switch strings.ToLower(strings.TrimSpace(input)) {
//		case "", "y", "yes":
//			continue
//		case "q", "quit", "n", "no":
//			return nil
//		default:
//			fmt.Fprintln(cmd.OutOrStdout(), "unrecognized input, press Enter to capture again or q to quit")
//		}
//	}
//}
//
//func captureAndPrintOCR(match windowMatch, tempDir string, renderer ocrOutputRenderer, captureNumber int) error {
//	if !ocrGameNoFocus {
//		if err := focusWindow(match.Client); err != nil {
//			return err
//		}
//		time.Sleep(150 * time.Millisecond)
//	}
//
//	var region string
//	var err error
//	if ocrGameWindowOnly {
//		region, err = windowRegion(match.Client)
//		if err != nil {
//			return err
//		}
//	} else {
//		region, err = selectCaptureRegion()
//		if err != nil {
//			return err
//		}
//	}
//
//	imagePath := filepath.Join(tempDir, fmt.Sprintf("capture-%d.png", time.Now().UnixNano()))
//	if err := captureRegion(region, imagePath); err != nil {
//		return err
//	}
//
//	text, err := runTesseractOCR(imagePath)
//	if err != nil {
//		return err
//	}
//
//	renderer.PrintResult(captureNumber, strings.TrimSpace(text))
//	return nil
//}
//
//func (r ocrOutputRenderer) PrintResult(captureNumber int, text string) {
//	var builder strings.Builder
//	if r.replace {
//		builder.WriteString("\033[2J\033[H")
//	}
//	builder.WriteString(fmt.Sprintf("\n=== OCR %d ===\n", captureNumber))
//	if text == "" {
//		builder.WriteString("(no text recognized)\n")
//	} else {
//		builder.WriteString(text)
//		if !strings.HasSuffix(text, "\n") {
//			builder.WriteByte('\n')
//		}
//	}
//	fmt.Fprint(r.writer, builder.String())
//}
//
//func focusWindow(client hyprClient) error {
//	if strings.TrimSpace(client.Address) == "" {
//		return nil
//	}
//	cmd := exec.Command("hyprctl", "dispatch", "focuswindow", "address:"+client.Address)
//	if output, err := cmd.CombinedOutput(); err != nil {
//		return fmt.Errorf("focus matched window: %w: %s", err, strings.TrimSpace(string(output)))
//	}
//	return nil
//}
//
//func selectCaptureRegion() (string, error) {
//	cmd := exec.Command("slurp")
//	output, err := cmd.CombinedOutput()
//	if err != nil {
//		return "", fmt.Errorf("select capture region: %w", err)
//	}
//
//	region := strings.TrimSpace(string(output))
//	if region == "" {
//		return "", errors.New("capture region was empty")
//	}
//	return region, nil
//}
//
//func windowRegion(client hyprClient) (string, error) {
//	if len(client.At) < 2 || len(client.Size) < 2 {
//		return "", errors.New("matched Hyprland window did not include usable geometry")
//	}
//
//	x := client.At[0]
//	y := client.At[1]
//	width := client.Size[0]
//	height := client.Size[1]
//	if width <= 0 || height <= 0 {
//		return "", errors.New("matched Hyprland window reported an invalid size")
//	}
//
//	return fmt.Sprintf("%d,%d %dx%d", x, y, width, height), nil
//}
//
//func captureRegion(region, imagePath string) error {
//	cmd := exec.Command("grim", "-g", region, imagePath)
//	if output, err := cmd.CombinedOutput(); err != nil {
//		return fmt.Errorf("capture screenshot: %w: %s", err, strings.TrimSpace(string(output)))
//	}
//	return nil
//}
//
//func runTesseractOCR(imagePath string) (string, error) {
//	format, err := normalizedOCRFormat()
//	if err != nil {
//		return "", err
//	}
//
//	args := []string{
//		imagePath,
//		"stdout",
//		"-l", ocrGameLang,
//		"--psm", strconv.Itoa(ocrGamePSM),
//	}
//	if format != "text" {
//		args = append(args, format)
//	}
//
//	cmd := exec.Command("tesseract", args...)
//	var stdout bytes.Buffer
//	var stderr bytes.Buffer
//	cmd.Stdout = &stdout
//	cmd.Stderr = &stderr
//	if err := cmd.Run(); err != nil {
//		return "", fmt.Errorf("run tesseract: %w: %s", err, strings.TrimSpace(stderr.String()))
//	}
//	return stdout.String(), nil
//}
